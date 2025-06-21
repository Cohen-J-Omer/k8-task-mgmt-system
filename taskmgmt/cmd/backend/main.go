package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/proto"
	"github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/internal/config"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type server struct {
	pb.UnimplementedTaskServiceServer
	mongoCol *mongo.Collection
}

func (s *server) CreateTask(ctx context.Context, req *pb.Task) (*pb.Task, error) {
    req.Id = uuid.New().String() // Generate a new UUID for the task ID
    _, err := s.mongoCol.InsertOne(ctx, req)
    return req, err
}

func (s *server) GetTask(ctx context.Context, req *pb.TaskID) (*pb.Task, error) {
    var task pb.Task
    err := s.mongoCol.FindOne(ctx, bson.M{"id": req.Id}).Decode(&task)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "task with id %s not found", req.Id)
    }
    return &task, nil
}

func (s *server) GetTasks(ctx context.Context, _ *pb.Empty) (*pb.TaskList, error) {
    cursor, err := s.mongoCol.Find(ctx, bson.M{})
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
    }
    defer cursor.Close(ctx)
    var tasks []*pb.Task
    for cursor.Next(ctx) {
        var t pb.Task
        if err := cursor.Decode(&t); err == nil {
            tasks = append(tasks, &t)
        }
    }
    // Always return a TaskList, possibly empty
    return &pb.TaskList{Tasks: tasks}, nil
}

// update task implementing upsert behavior, i.e., 
// create a new task with that ID if it does not exist, or update it if it does
func (s *server) UpdateTask(ctx context.Context, req *pb.Task) (*pb.Task, error) {
    filter := bson.M{"id": req.Id}
    update := bson.M{"$set": req}
    opts := options.Update().SetUpsert(true)
    _, err := s.mongoCol.UpdateOne(ctx, filter, update, opts)
    return req, err
}

func (s *server) DeleteTask(ctx context.Context, req *pb.TaskID) (*pb.Task, error) {
    var deletedTask pb.Task
    err := s.mongoCol.FindOne(ctx, bson.M{"id": req.Id}).Decode(&deletedTask)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "task with id %s not found", req.Id)
    }
    _, err = s.mongoCol.DeleteOne(ctx, bson.M{"id": req.Id})
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to delete task: %v", err)
    }
    return &deletedTask, nil
}


func main() {
	// loads .env for local debugging
	debug:=config.LoadDotenvIfDebug()

	mongoUser, okUser := os.LookupEnv("MONGO_USERNAME")
	mongoPass, okPass := os.LookupEnv("MONGO_PASSWORD")

	if !okUser || mongoUser == "" {
		log.Fatal("Environment variable MONGO_USERNAME is not set")
	}
	if !okPass || mongoPass == "" {
		log.Fatal("Environment variable MONGO_PASSWORD is not set")
	}

	mongoHost := "mongodb" // default for Kubernetes
	// replace host with localhost for local debugging
	if debug{
		mongoHost = "localhost"
	}
	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:27017", mongoUser, mongoPass, mongoHost)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	col := client.Database("tasks").Collection("tasks")

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	// register task service
	grpcServer := grpc.NewServer()
	pb.RegisterTaskServiceServer(grpcServer, &server{mongoCol: col})

	// Register gRPC health check service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Graceful shutdown
	go func() {
		log.Println("gRPC backend listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gRPC backend...")

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("gRPC backend exited gracefully")
	case <-time.After(10 * time.Second):
		log.Println("gRPC backend forced to stop")
		grpcServer.Stop()
	}
}
