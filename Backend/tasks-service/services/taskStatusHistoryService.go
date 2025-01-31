package services

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"tasks-service/model"
	"tasks-service/repositories"
)

type TaskStatusHistoryService struct {
	repo   *repositories.TaskStatusHistoryRepo
	logger *log.Logger
}

func NewTaskStatusHistoryService(repo *repositories.TaskStatusHistoryRepo, logger *log.Logger) *TaskStatusHistoryService {
	return &TaskStatusHistoryService{
		repo:   repo,
		logger: logger,
	}
}

func (s *TaskStatusHistoryService) InsertStatusHistory(ctx context.Context, taskID primitive.ObjectID, status model.Status) (*model.TaskStatusHistory, error) {
	ctx, span := otel.Tracer("tasks-service").Start(ctx, "InsertStatusHistoryService")
	defer span.End()

	history := &model.TaskStatusHistory{
		TaskID:    taskID,
		Status:    status,
		Timestamp: time.Now(),
	}

	span.SetAttributes(
		attribute.String("task.id", taskID.Hex()),
		attribute.String("status", string(status)),
	)

	_, err := s.repo.Insert(ctx, history)
	if err != nil {
		span.RecordError(err)
		s.logger.Println("Failed to insert task status history:", err)
		return nil, err
	}

	span.SetAttributes(attribute.String("task_status_history.inserted_id", history.ID.Hex()))
	return history, nil
}

func (s *TaskStatusHistoryService) GetStatusHistoryByTaskID(ctx context.Context, taskID primitive.ObjectID) ([]model.TaskStatusHistory, error) {
	
	log.Println("Debug: GetStatusHistoryByTaskID Service Triggered")

	ctx, span := otel.Tracer("tasks-service").Start(ctx, "GetStatusHistoryByTaskIDService")
	defer span.End()

	span.SetAttributes(attribute.String("task.id", taskID.Hex()))

	records, err := s.repo.GetByTaskId(ctx, taskID)
	if err != nil {
		span.RecordError(err)
		s.logger.Println("Failed to get task status history by TaskID:", err)
		return nil, err
	}

	log.Println("Debug: GetStatusHistoryByTaskID Service Finished")

	span.SetAttributes(attribute.Int("history.count", len(records)))
	return records, nil
}
