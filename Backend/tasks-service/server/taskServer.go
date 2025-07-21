package server

import (
	"context"
	"fmt"
	taskpb "pb/taskpb"
	"tasks-service/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	ctx, span := otel.Tracer("tasks-service").Start(ctx, "GetUnassignedTasks")
	defer span.End()

	span.SetAttributes(attribute.String("project.id", req.GetProjectId()))

	projectID, err := primitive.ObjectIDFromHex(req.GetProjectId())
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid project ID: %v", err)
	}

	tasks, err := s.taskRepo.GetUnassignedTasks(ctx, projectID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("error fetching unassigned tasks: %v", err)
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))

	var taskResponses []*taskpb.Task
	for _, task := range tasks {
		var member *wrapperspb.StringValue
		if !task.Member.IsZero() {
			member = wrapperspb.String(task.Member.Hex())
		}

		taskResponses = append(taskResponses, &taskpb.Task{
			Id:          task.ID.Hex(),
			Title:       task.Title,
			Description: task.Description,
			Status:      string(task.Status),
			Project:     task.Project.Hex(),
			Member:      member,
			Blocked:     task.Blocked,
		})
	}

	return &taskpb.TaskResponse{Tasks: taskResponses}, nil
}

func (s *TaskServer) CheckMemberTasksInProgress(ctx context.Context, req *taskpb.MemberRequest) (*taskpb.BoolResponse, error) {
	ctx, span := otel.Tracer("tasks-service").Start(ctx, "CheckMemberTasksInProgress")
	defer span.End()

	span.SetAttributes(attribute.String("member.id", req.GetMemberId()))

	memberID, err := primitive.ObjectIDFromHex(req.GetMemberId())
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid member ID: %v", err)
	}

	hasTasks, err := s.taskRepo.HasPendingOrInProgressTasks(ctx, memberID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("error checking in-progress tasks: %v", err)
	}

	span.SetAttributes(attribute.Bool("has.inProgress.tasks", hasTasks))

	return &taskpb.BoolResponse{Value: hasTasks}, nil
}
