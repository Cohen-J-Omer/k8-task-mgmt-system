package validator

import (
	"errors"

	pb "github.com/Cohen-J-Omer/k8-task-mgmt-system/taskmgmt/proto"
)

func ValidateTask(task *pb.Task) error {
	if task.Title == "" {
		return errors.New("title is required")
	}
	return nil
}
