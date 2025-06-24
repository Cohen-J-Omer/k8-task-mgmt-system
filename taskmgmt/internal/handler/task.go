package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/internal/validator"
	pb "github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)
// implements grpc's TaskServiceClient 
// TaskHandler handles API HTTP requests for task management
// using a gRPC client to communicate with the backend service.
type TaskHandler struct {
	client pb.TaskServiceClient
}

func NewTaskHandler(client pb.TaskServiceClient) *TaskHandler {
	return &TaskHandler{client: client}
}

// CreateTask handles the creation of a new task.
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req pb.Task
	// bind the JSON body to the task struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// validate task object using the validator package
	if err := validator.ValidateTaskCreate(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Set a timeout context for the gRPC call
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// Create the task using the gRPC client
	resp, err := h.client.CreateTask(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetTasks retrieves a list of tasks from the backend service.
func (h *TaskHandler) GetTasks(c *gin.Context) {
	// Call the gRPC service to get the list of tasks
	// not using a timeout here, as it may take longer to fetch tasks
	taskList, err := h.client.GetTasks(c.Request.Context(), &pb.Empty{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
        return
    }
	// Ensure tasks is an array, not nil
    if taskList.Tasks == nil {
        taskList.Tasks = []*pb.Task{}
    }
    c.JSON(http.StatusOK, taskList)
}

// GetTask retrieves a specific task by its ID.
func (h *TaskHandler) GetTask(c *gin.Context) {
    id := c.Param("id") // Extract the task ID from the URL parameter
	// Validate that the ID is not empty
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task ID is required"})
		return
	}
    req := &pb.TaskID{Id: id}

	// Set a timeout context for the gRPC call
    ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
    defer cancel()

    task, err := h.client.GetTask(ctx, req)
    if err != nil {
        status, ok := status.FromError(err)
		// Check if the error is a NotFound error
        if ok && status.Code() == codes.NotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("task with id %s not found", id)})
            return
        }
		// For other errors, return a generic internal server error
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task"})
        return
    }
    c.JSON(http.StatusOK, task)
}

// UpdateTask updates an existing task by its ID.
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	// Validate that the ID is not empty
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task ID is required"})
		return
	}
	var req pb.Task
	// bind the JSON body to the task struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Id = id
	// validate task object using the validator package
	if err := validator.ValidateTaskCreate(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Set a timeout context for the gRPC call
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// Update the task using the gRPC client
	resp, err := h.client.UpdateTask(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteTask deletes a specific task by its ID.
func (h *TaskHandler) DeleteTask(c *gin.Context) {
    id := c.Param("id")
	// Validate that the ID is not empty
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task ID is required"})
		return
	}
    req := &pb.TaskID{Id: id}

	// Set a timeout context for the gRPC call
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

    deletedTask, err := h.client.DeleteTask(ctx, req)
    if err != nil {
        status, ok := status.FromError(err)
		// Check if the error is a NotFound error
        if ok && status.Code() == codes.NotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("task with id %s not found", id)})
            return
        }
		// For other errors, return a generic internal server error
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete task"})
        return
    }
    c.JSON(http.StatusOK, deletedTask)
}
