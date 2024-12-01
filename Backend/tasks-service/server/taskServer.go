package server

import (
	"context"
	"fmt"
	taskpb "pb/taskpb"
	"tasks-service/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TaskServer struct implements the TaskServiceServer interface from the proto
type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	taskRepo *repositories.TaskRepo
}

// NewTaskServer creates a new instance of TaskServer
func NewTaskServer(taskRepo *repositories.TaskRepo) *TaskServer {
	return &TaskServer{
		taskRepo: taskRepo,
	}
}

// GetUnassignedTasks implements the GetUnassignedTasks gRPC method
func (s *TaskServer) GetUnassignedTasks(ctx context.Context, req *taskpb.ProjectRequest) (*taskpb.TaskResponse, error) {
	// Convert project_id string to primitive.ObjectID
	projectID, err := primitive.ObjectIDFromHex(req.GetProjectId())
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %v", err)
	}

	// Fetch tasks that are unassigned (i.e., member is empty or zero)
	tasks, err := s.taskRepo.GetUnassignedTasks(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error fetching unassigned tasks: %v", err)
	}

	// Convert tasks to gRPC response format
	var taskResponses []*taskpb.Task
	for _, task := range tasks {
		// Convert Member to *wrapperspb.StringValue, which is nullable in your proto file
		var member *wrapperspb.StringValue
		if !task.Member.IsZero() {
			member = wrapperspb.String(task.Member.Hex()) // Wrap ObjectID as string
		}

		taskResponses = append(taskResponses, &taskpb.Task{
			Id:          task.ID.Hex(),
			Title:       task.Title,
			Description: task.Description,
			Status:      string(task.Status),
			Project:     task.Project.Hex(), // Assuming project is stored as ObjectID
			Member:      member,             // Nullable member field
			Blocked:     task.Blocked,
		})
	}

	// Return the tasks in the gRPC response
	return &taskpb.TaskResponse{Tasks: taskResponses}, nil
}
