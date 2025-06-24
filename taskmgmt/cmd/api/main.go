package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/internal/config"
	"github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/internal/handler"
	"github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/internal/middleware"
	pb "github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// loads .env for local debugging
	config.LoadDotenvIfDebug()

	grpcAddr, ok := os.LookupEnv("BACKEND_GRPC_ADDR")
	if !ok || grpcAddr == "" {
		log.Fatal("Environment variable BACKEND_GRPC_ADDR is not set")
	}

	bearerToken, ok2 := os.LookupEnv("BEARER_TOKEN")
	if !ok2 || bearerToken == "" {
		log.Fatal("Environment variable BEARER_TOKEN is not set")
	}
	// Create a gRPC client connection
	log.Printf("Connecting to gRPC server at %s", grpcAddr)
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create a new gRPC client
	client := pb.NewTaskServiceClient(conn)

	// Set up Gin router
	r := gin.Default()
	
	taskHandler := handler.NewTaskHandler(client)
	// add a health readiness/liveness entry point for k8
	// This allows Kubernetes HPA to check the health of the API server.
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	// /health is always accessible not requiring authentication token
	r.Use(middleware.AuthMiddleware(bearerToken))
	r.POST("/tasks", taskHandler.CreateTask)
	r.GET("/tasks", taskHandler.GetTasks)
	r.GET("/tasks/:id", taskHandler.GetTask)
	r.PUT("/tasks/:id", taskHandler.UpdateTask)
	r.DELETE("/tasks/:id", taskHandler.DeleteTask)
	
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	// Start the API HTTP server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("REST API server started on :8080")
		
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down REST API server...")
	
	// gives the API server up to 10 seconds to finish handling any in-flight requests and shut down gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("REST API forced to shutdown:", err)
	}
	log.Println("REST API server exited gracefully")
}
