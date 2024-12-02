package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"projects-service/model"
	"projects-service/repositories"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProjectService struct {
	repo      *repositories.ProjectRepo
	publisher *nats.Conn
}

func NewProjectService(repo *repositories.ProjectRepo, nc *nats.Conn) *ProjectService {
	return &ProjectService{
		repo:      repo,
		publisher: nc,
	}
}

func (ps *ProjectService) GetAllProjects(ctx context.Context) ([]model.Project, error) {
	return ps.repo.GetAll(ctx)
}

func (ps *ProjectService) GetProjectByName(ctx context.Context, name string) (*model.Project, error) {
	return ps.repo.GetByName(ctx, name)
}

func (ps *ProjectService) GetProjectById(ctx context.Context, id primitive.ObjectID) (*model.Project, error) {
	if id.IsZero() {
		return nil, errors.New("invalid project ID")
	}

	project, err := ps.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}
	return project, nil
}

func (ps *ProjectService) CreateProject(ctx context.Context, project *model.Project) (*mongo.InsertOneResult, error) {
	if project == nil {
		return nil, errors.New("project cannot be nil")
	}
	if project.Name == "" {
		return nil, errors.New("project name is required")
	}
	if project.MinMembers <= 0 {
		return nil, errors.New("minimum members must be greater than 0")
	}
	if project.MaxMembers < project.MinMembers {
		return nil, errors.New("maximum members cannot be less than minimum members")
	}
	if len(project.Members) < project.MinMembers || len(project.Members) > project.MaxMembers {
		return nil, errors.New("number of members must be between the minimum and maximum limits")
	}

	return ps.repo.Insert(ctx, project)
}

func (ps *ProjectService) UpdateProject(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*mongo.UpdateResult, error) {
	if id.IsZero() {
		return nil, errors.New("invalid project ID")
	}
	if len(updateData) == 0 {
		return nil, errors.New("update data cannot be empty")
	}

	return ps.repo.Update(ctx, id, updateData)
}

func (ps *ProjectService) DeleteProjectWithTasks(ctx context.Context, projectID primitive.ObjectID) error {
	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)
	logger.Printf("Starting to delete project with ID: %s", projectID.Hex())

	_, err := ps.GetProjectById(ctx, projectID)
	if err != nil {
		logger.Printf("Project not found: %v", err)
		return err
	}

	responseCh := make(chan *nats.Msg, 1)
	sub, err := ps.publisher.ChanSubscribe("tasks.deleted", responseCh)
	if err != nil {
		logger.Printf("Failed to subscribe to confirmation channel: %v", err)
		return fmt.Errorf("failed to subscribe to confirmation channel: %w", err)
	}
	defer sub.Unsubscribe()

	// Ensure subscription is ready
	err = ps.publisher.Flush()
	if err != nil {
		logger.Printf("NATS flush error: %v", err)
		return err
	}

	logger.Printf("Subscribed to tasks.deleted and ready to publish tasks.delete event")

	payload, _ := json.Marshal(map[string]string{"project_id": projectID.Hex()})
	err = ps.publisher.Publish("tasks.delete", payload)
	if err != nil {
		logger.Printf("Failed to publish delete tasks event: %v", err)
		return fmt.Errorf("failed to publish delete tasks event: %w", err)
	}
	logger.Printf("Published tasks.delete event for project ID: %s", projectID.Hex())

	select {
	case msg := <-responseCh:
		logger.Println("Received message from tasks.deleted topic")
		var event map[string]string
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logger.Printf("Invalid confirmation message format: %v", err)
			return fmt.Errorf("invalid confirmation message format: %w", err)
		}

		if event["project_id"] != projectID.Hex() {
			logger.Printf("Confirmation received for incorrect project ID: %s", event["project_id"])
			return fmt.Errorf("confirmation received for incorrect project ID")
		}

		logger.Printf("Successfully received confirmation for project ID: %s", projectID.Hex())
		_, err = ps.repo.Delete(ctx, projectID)
		if err != nil {
			logger.Printf("Failed to delete project: %v", err)
			return fmt.Errorf("failed to delete project: %w", err)
		}

	case <-time.After(20 * time.Second):
		logger.Printf("Timeout waiting for task deletion confirmation for project ID: %s", projectID.Hex())
		return fmt.Errorf("timeout waiting for task deletion confirmation")
	}

	logger.Printf("Successfully deleted project with ID: %s", projectID.Hex())
	return nil
}

func (ps *ProjectService) GetProjectsByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.Project, error) {
	return ps.repo.GetProjectsByUserID(ctx, userID)
}

func (ps *ProjectService) GetProjectsByManagerID(ctx context.Context, userID primitive.ObjectID) ([]model.Project, error) {
	return ps.repo.GetProjectsByManagerID(ctx, userID)
}