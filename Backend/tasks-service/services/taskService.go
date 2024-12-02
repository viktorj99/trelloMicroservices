package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"tasks-service/client"
	"tasks-service/model"
	"tasks-service/repositories"
)

type TaskService struct {
	repo       *repositories.TaskRepo
	natsClient *client.NATSClient
	logger     *log.Logger
}

func NewTaskService(repo *repositories.TaskRepo, natsClient *client.NATSClient, logger *log.Logger) *TaskService {
	return &TaskService{
		repo:       repo,
		natsClient: natsClient,
		logger:     logger,
	}
}

func (s *TaskService) Start() {
	s.logger.Println("Starting TaskService...")
	go s.ListenForTaskDeletion()
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

// DeleteTasksForProject deletes all tasks associated with a project when a delete event is received from NATS.
func (s *TaskService) DeleteTasksForProject(ctx context.Context, projectID primitive.ObjectID) error {
	s.logger.Printf("Starting to delete tasks for project ID: %s", projectID.Hex())

	tasks, err := s.repo.GetByProjectId(ctx, projectID)
	if err != nil {
		s.logger.Printf("Failed to find tasks for project: %v", err)
		return err
	}
	s.logger.Printf("Found %d tasks for project ID: %s", len(tasks), projectID.Hex())

	for _, task := range tasks {
		s.logger.Printf("Deleting task with ID: %s", task.ID.Hex())
		_, err := s.repo.DeleteTasks(ctx, task.ID)
		if err != nil {
			s.logger.Printf("Failed to delete task with ID %s: %v", task.ID.Hex(), err)
			return err
		}
	}
	s.logger.Printf("Successfully deleted all tasks for project ID: %s", projectID.Hex())

	confirmation := map[string]string{"project_id": projectID.Hex()}
	confirmationData, err := json.Marshal(confirmation)
	if err != nil {
		s.logger.Printf("Failed to marshal confirmation data: %v", err)
		return err
	}

	err = s.natsClient.Conn.Publish("tasks.deleted", confirmationData)
	if err != nil {
		s.logger.Printf("Failed to publish task deletion confirmation: %v", err)
		return err
	}

	s.logger.Printf("Published task deletion confirmation for project ID: %s", projectID.Hex())
	return nil
}

// ListenForTaskDeletion listens for the "tasks.delete" event from NATS and processes it.
func (s *TaskService) ListenForTaskDeletion() {
	s.logger.Println("Starting to listen for tasks.delete events")

	_, err := s.natsClient.Conn.Subscribe("tasks.delete", func(msg *nats.Msg) {
		s.logger.Println("Received tasks.delete event")
		var event map[string]string
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			s.logger.Printf("Error unmarshalling NATS message: %v", err)
			return
		}

		projectIDStr, ok := event["project_id"]
		if !ok {
			s.logger.Println("Missing project_id in NATS message")
			return
		}

		projectID, err := primitive.ObjectIDFromHex(projectIDStr)
		if err != nil {
			s.logger.Printf("Invalid project ID format: %v", err)
			return
		}

		s.logger.Printf("Deleting tasks for project ID: %s", projectID.Hex())
		err = s.DeleteTasksForProject(context.Background(), projectID)
		if err != nil {
			s.logger.Printf("Failed to delete tasks for project %s: %v", projectID.Hex(), err)
		}
	})
	if err != nil {
		s.logger.Fatalf("Error subscribing to NATS task deletion event: %v", err)
	}
}
