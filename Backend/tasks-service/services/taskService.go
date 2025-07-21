package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

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
	ctx, span := otel.Tracer("tasks-service").Start(ctx, "GetAllTasksService")
	defer span.End()

	tasks, err := s.repo.GetAll(ctx)
	if err != nil {
		span.RecordError(err)
		s.logger.Println("Failed to get all tasks:", err)
		return nil, err
	}
	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))
	return tasks, nil
}

func (s *TaskService) GetTaskById(ctx context.Context, id string) (*model.Task, error) {
	ctx, span := otel.Tracer("tasks-service").Start(ctx, "GetTaskByIdService")
	defer span.End()
	span.SetAttributes(attribute.String("task.id", id))

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		span.RecordError(err)
		s.logger.Println("Invalid task ID format:", err)
		return nil, err
	}

	task, err := s.repo.GetById(ctx, objectID)
	if err != nil {
		span.RecordError(err)
		s.logger.Println("Failed to get task by ID:", err)
		return nil, err
	}
	return task, nil
}

func (s *TaskService) CreateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	ctx, span := otel.Tracer("tasks-service").Start(ctx, "CreateTaskService")
	defer span.End()

	span.SetAttributes(attribute.String("task.title", task.Title), attribute.String("task.status", string(task.Status)))

	if task.Status != "PENDING" {
		err := errors.New("task status must be 'PENDING'")
		span.RecordError(err)
		s.logger.Println(err)
		return nil, err
	}

	task.ID = primitive.NewObjectID()

	_, err := s.repo.Insert(ctx, task)
	if err != nil {
		span.RecordError(err)
		s.logger.Println("Failed to create task:", err)
		return nil, err
	}
	return task, nil
}

func (s *TaskService) GetTasksByProjectId(ctx context.Context, projectId string) ([]model.Task, error) {
	ctx, span := otel.Tracer("tasks-service").Start(ctx, "GetTasksByProjectIdService")
	defer span.End()

	span.SetAttributes(attribute.String("project.id", projectId))

	objectID, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		span.RecordError(err)
		s.logger.Println("Invalid project ID format:", err)
		return nil, err
	}

	tasks, err := s.repo.GetByProjectId(ctx, objectID)
	if err != nil {
		span.RecordError(err)
		s.logger.Println("Failed to get tasks by project ID:", err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))
	return tasks, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id primitive.ObjectID, updateData bson.M) error {
	ctx, span := otel.Tracer("tasks-service").Start(ctx, "UpdateTaskService")
	defer span.End()

	span.SetAttributes(attribute.String("task.id", id.Hex()), attribute.String("update.fields", fmt.Sprintf("%v", updateData)))

	_, err := s.repo.Update(ctx, id, updateData)
	if err != nil {
		span.RecordError(err)
		s.logger.Println("Failed to update task:", err)
		return err
	}

	span.SetAttributes(attribute.Bool("update.success", true))
	return nil
}

func (s *TaskService) GetTasksWithDependencies(ctx context.Context, projectId string) ([]model.TaskWithDependencies, error) {
	return s.repo.GetTasksWithDependencies(ctx, projectId)
}

// DeleteTasksForProject deletes all tasks associated with a project when a delete event is received from NATS.
func (s *TaskService) DeleteTasksForProject(ctx context.Context, projectID primitive.ObjectID) error {
	ctx, span := otel.Tracer("tasks-service").Start(ctx, "DeleteTasksForProject")
	defer span.End()

	span.SetAttributes(attribute.String("project.id", projectID.Hex()))

	s.logger.Printf("Starting to delete tasks for project ID: %s", projectID.Hex())

	tasks, err := s.repo.GetByProjectId(ctx, projectID)
	if err != nil {
		span.RecordError(err)
		s.logger.Printf("Failed to find tasks for project: %v", err)
		return err
	}

	span.SetAttributes(attribute.Int("tasks.found", len(tasks)))

	for _, task := range tasks {
		s.logger.Printf("Deleting task with ID: %s", task.ID.Hex())
		_, err := s.repo.DeleteTasks(ctx, task.ID)
		if err != nil {
			span.RecordError(err)
			s.logger.Printf("Failed to delete task with ID %s: %v", task.ID.Hex(), err)
			return err
		}
	}

	span.SetAttributes(attribute.Bool("deletion.success", true))
	s.logger.Printf("Successfully deleted all tasks for project ID: %s", projectID.Hex())

	confirmation := map[string]string{"project_id": projectID.Hex()}
	confirmationData, err := json.Marshal(confirmation)
	if err != nil {
		span.RecordError(err)
		s.logger.Printf("Failed to marshal confirmation data: %v", err)
		return err
	}

	err = s.natsClient.Conn.Publish("tasks.deleted", confirmationData)
	if err != nil {
		span.RecordError(err)
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
		ctx, span := otel.Tracer("tasks-service").Start(context.Background(), "ListenForTaskDeletion")
		defer span.End()

		s.logger.Println("Received tasks.delete event")

		var event map[string]string
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			span.RecordError(err)
			s.logger.Printf("Error unmarshalling NATS message: %v", err)
			return
		}

		projectIDStr, ok := event["project_id"]
		if !ok {
			err := errors.New("missing project_id in NATS message")
			span.RecordError(err)
			s.logger.Println(err)
			return
		}

		span.SetAttributes(attribute.String("project.id", projectIDStr))

		projectID, err := primitive.ObjectIDFromHex(projectIDStr)
		if err != nil {
			span.RecordError(err)
			s.logger.Printf("Invalid project ID format: %v", err)
			return
		}

		s.logger.Printf("Deleting tasks for project ID: %s", projectID.Hex())
		err = s.DeleteTasksForProject(ctx, projectID)
		if err != nil {
			span.RecordError(err)
			s.logger.Printf("Failed to delete tasks for project %s: %v", projectID.Hex(), err)
		}
	})
	if err != nil {
		s.logger.Fatalf("Error subscribing to NATS task deletion event: %v", err)
	}
}
