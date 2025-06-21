package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/internal/validator"
	pb "github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

type TaskHandler struct {
	client pb.TaskServiceClient
}

func NewTaskHandler(client pb.TaskServiceClient) *TaskHandler {
	return &TaskHandler{client: client}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req pb.Task
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateTask(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := h.client.CreateTask(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
    taskList, err := h.client.GetTasks(context.Background(), &pb.Empty{})
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

func (h *TaskHandler) GetTask(c *gin.Context) {
    id := c.Param("id")
    req := &pb.TaskID{Id: id}
    task, err := h.client.GetTask(context.Background(), req)
    if err != nil {
        st, ok := status.FromError(err)
        if ok && st.Code() == codes.NotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("task with id %s not found", id)})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task"})
        return
    }
    c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var req pb.Task
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Id = id
	if err := validator.ValidateTask(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := h.client.UpdateTask(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
    id := c.Param("id")
    req := &pb.TaskID{Id: id}
    deletedTask, err := h.client.DeleteTask(context.Background(), req)
    if err != nil {
        st, ok := status.FromError(err)
        if ok && st.Code() == codes.NotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("task with id %s not found", id)})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete task"})
        return
    }
    c.JSON(http.StatusOK, deletedTask)
}
