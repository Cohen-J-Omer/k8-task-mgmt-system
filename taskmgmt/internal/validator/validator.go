package validator

import (
	"errors"

	pb "github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/proto"
)

// ValidateTaskCreate validates the task creation/update request.
// It checks that the title and description are not empty and within length limits.
func ValidateTaskCreate(task *pb.Task) error {
	if task.Title == "" {
		return errors.New("title is required")
	}
	if len(task.Title) > 100 {
		return errors.New("title must be at most 100 characters")
	}
	if task.Description == "" {
		return errors.New("description is required")
	}
	if len(task.Description) > 1000 {
		return errors.New("description must be at most 1000 characters")
	}
	return nil
}