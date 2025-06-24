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

// implements gRPC's TaskServiceServer interface
// This server handles gRPC requests for task management.
// It connects to a MongoDB database to store and retrieve tasks.
type server struct {
	pb.UnimplementedTaskServiceServer
	mongoCol *mongo.Collection  // collection handler for the "tasks" MongoDB collection
}

// CreateTask creates a new task in the MongoDB collection.
func (s *server) CreateTask(ctx context.Context, req *pb.Task) (*pb.Task, error) {
    req.Id = uuid.New().String() // Generate a new UUID for the task ID
    _, err := s.mongoCol.InsertOne(ctx, req)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to create task: %v", err)
    }
    return req, nil
}

// GetTask retrieves a task by its ID from the MongoDB collection.
func (s *server) GetTask(ctx context.Context, req *pb.TaskID) (*pb.Task, error) {
    var task pb.Task
    err := s.mongoCol.FindOne(ctx, bson.M{"id": req.Id}).Decode(&task)
	// If the task is not found, return a NotFound error
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "task with id %s not found", req.Id)
    }
    return &task, nil
}

// GetTasks retrieves all tasks from the MongoDB collection.
func (s *server) GetTasks(ctx context.Context, _ *pb.Empty) (*pb.TaskList, error) {
    cursor, err := s.mongoCol.Find(ctx, bson.M{})
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
    }
    defer cursor.Close(ctx)
    var tasks []*pb.Task
    for cursor.Next(ctx) {
        var t pb.Task
        if err := cursor.Decode(&t); err != nil {
			// Log the error but continue processing other tasks
			log.Printf("failed to decode task: %v", err)
		}else{
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
    opts := options.Update().SetUpsert(true)  // Enable upsert behavior, i.e., create if not exists
    _, err := s.mongoCol.UpdateOne(ctx, filter, update, opts)
	if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
    }
    return req, nil
}

// DeleteTask deletes a task by its ID from the MongoDB collection.
// It returns the deleted task if found, or an error if not found.
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
	// Connect to MongoDB using the provided URI
	log.Printf("Connecting to MongoDB at %s", mongoURI)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	col := client.Database("tasks").Collection("tasks") // Use/create the "tasks" collection in the "tasks" database

	// creates a TCP network listener on port 50051 for gRPC server 
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	// register server as a gRPC TaskServiceServer 
	grpcServer := grpc.NewServer()
	pb.RegisterTaskServiceServer(grpcServer, &server{mongoCol: col})

	// Register gRPC health check service for k8 readiness and liveness probes
	// This allows Kubernetes HPA to check the health of the gRPC server.
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Start the gRPC backend server
	go func() {
		log.Println("gRPC backend listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		}()
		
	// Wait for a termination signal (SIGINT or SIGTERM) to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gRPC backend...")

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()
	
	// Wait for the server to stop gracefully or timeout after 10 seconds
	select {
	case <-done:
		log.Println("gRPC backend exited gracefully")
	case <-time.After(10 * time.Second):
		log.Println("gRPC backend forced to stop")
		grpcServer.Stop()
	}
}
