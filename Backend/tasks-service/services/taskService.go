package services

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"tasks-service/model"
	"tasks-service/repositories"
)

type TaskService struct {
	repo   *repositories.TaskRepo
	logger *log.Logger
}

func NewTaskService(repo *repositories.TaskRepo, logger *log.Logger) *TaskService {
	return &TaskService{
		repo:   repo,
		logger: logger,
	}
}

func (s *TaskService) GetAllTasks(ctx context.Context) ([]model.Task, error) {
	tasks, err := s.repo.GetAll(ctx)
	if err != nil {
		s.logger.Println("Failed to get all tasks:", err)
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) GetTaskById(ctx context.Context, id string) (*model.Task, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		s.logger.Println("Invalid task ID format:", err)
		return nil, err
	}

	task, err := s.repo.GetById(ctx, objectID)
	if err != nil {
		s.logger.Println("Failed to get task by ID:", err)
		return nil, err
	}

	return task, nil
}

func (s *TaskService) CreateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	if task.Status != "PENDING" {
		err := errors.New("task status must be 'PENDING'")
		s.logger.Println(err)
		return nil, err
	}

	task.ID = primitive.NewObjectID()

	_, err := s.repo.Insert(ctx, task)
	if err != nil {
		s.logger.Println("Failed to create task:", err)
		return nil, err
	}

	return task, nil
}
