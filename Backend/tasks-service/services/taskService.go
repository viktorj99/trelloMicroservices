package services

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
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

func (s *TaskService) GetTasksByProjectId(ctx context.Context, projectId string) ([]model.Task, error) {
	objectID, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		s.logger.Println("Invalid project ID format:", err)
		return nil, err
	}

	tasks, err := s.repo.GetByProjectId(ctx, objectID)
	if err != nil {
		s.logger.Println("Failed to get tasks by project ID:", err)
		return nil, err
	}

	return tasks, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id primitive.ObjectID, updateData bson.M) error {
	_, err := s.repo.Update(ctx, id, updateData)
	if err != nil {
		s.logger.Println("Failed to update task:", err)
		return err
	}
	return nil
}
