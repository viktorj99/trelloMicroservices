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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	tracer := otel.Tracer("projects-service")
	ctx, span := tracer.Start(ctx, "GetAllProjectsService")
	defer span.End()

	projects, err := ps.repo.GetAll(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("projects.count", len(projects)))
	return projects, nil
}

func (ps *ProjectService) GetProjectByName(ctx context.Context, name string) (*model.Project, error) {
	tracer := otel.Tracer("projects-service")
	ctx, span := tracer.Start(ctx, "GetProjectByNameService")
	defer span.End()

	span.SetAttributes(attribute.String("project.name", name))

	project, err := ps.repo.GetByName(ctx, name)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return project, nil
}

func (ps *ProjectService) GetProjectById(ctx context.Context, id primitive.ObjectID) (*model.Project, error) {
	tracer := otel.Tracer("projects-service")
	ctx, span := tracer.Start(ctx, "GetProjectByIdService")
	defer span.End()

	span.SetAttributes(attribute.String("project.id", id.Hex()))

	if id.IsZero() {
		err := errors.New("invalid project ID")
		span.RecordError(err)
		return nil, err
	}

	project, err := ps.repo.GetById(ctx, id)
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			span.SetAttributes(attribute.Bool("db.result.empty", true))
			return nil, errors.New("project not found")
		}
		return nil, err
	}

	return project, nil
}

func (ps *ProjectService) CreateProject(ctx context.Context, project *model.Project) (*mongo.InsertOneResult, error) {
	tracer := otel.Tracer("projects-service")
	ctx, span := tracer.Start(ctx, "CreateProjectService")
	defer span.End()

	span.SetAttributes(attribute.String("project.name", project.Name))

	if project == nil {
		err := errors.New("project cannot be nil")
		span.RecordError(err)
		return nil, err
	}
	if project.Name == "" {
		err := errors.New("project name is required")
		span.RecordError(err)
		return nil, err
	}
	if project.MinMembers <= 0 {
		err := errors.New("minimum members must be greater than 0")
		span.RecordError(err)
		return nil, err
	}
	if project.MaxMembers < project.MinMembers {
		err := errors.New("maximum members cannot be less than minimum members")
		span.RecordError(err)
		return nil, err
	}
	if len(project.Members) < project.MinMembers || len(project.Members) > project.MaxMembers {
		err := errors.New("number of members must be between the minimum and maximum limits")
		span.RecordError(err)
		return nil, err
	}

	result, err := ps.repo.Insert(ctx, project)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.String("db.inserted_id", result.InsertedID.(primitive.ObjectID).Hex()))
	return result, nil
}

func (ps *ProjectService) UpdateProject(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*mongo.UpdateResult, error) {
	tracer := otel.Tracer("projects-service")
	ctx, span := tracer.Start(ctx, "UpdateProjectService")
	defer span.End()

	span.SetAttributes(attribute.String("project.id", id.Hex()))

	if id.IsZero() {
		err := errors.New("invalid project ID")
		span.RecordError(err)
		return nil, err
	}
	if len(updateData) == 0 {
		err := errors.New("update data cannot be empty")
		span.RecordError(err)
		return nil, err
	}

	result, err := ps.repo.Update(ctx, id, updateData)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return result, nil
}

func (ps *ProjectService) DeleteProjectWithTasks(ctx context.Context, projectID primitive.ObjectID) error {
	tracer := otel.Tracer("projects-service")
	ctx, span := tracer.Start(ctx, "DeleteProjectWithTasksService")
	defer span.End()

	span.SetAttributes(attribute.String("project.id", projectID.Hex()))

	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)
	logger.Printf("Starting to delete project with ID: %s", projectID.Hex())

	_, err := ps.GetProjectById(ctx, projectID)
	if err != nil {
		logger.Printf("Project not found: %v", err)
		span.RecordError(err)
		return err
	}

	responseCh := make(chan *nats.Msg, 1)
	sub, err := ps.publisher.ChanSubscribe("tasks.deleted", responseCh)
	if err != nil {
		logger.Printf("Failed to subscribe to confirmation channel: %v", err)
		span.RecordError(err)
		return fmt.Errorf("failed to subscribe to confirmation channel: %w", err)
	}
	defer sub.Unsubscribe()

	err = ps.publisher.Flush()
	if err != nil {
		logger.Printf("NATS flush error: %v", err)
		span.RecordError(err)
		return err
	}

	payload, _ := json.Marshal(map[string]string{"project_id": projectID.Hex()})
	err = ps.publisher.Publish("tasks.delete", payload)
	if err != nil {
		logger.Printf("Failed to publish delete tasks event: %v", err)
		span.RecordError(err)
		return fmt.Errorf("failed to publish delete tasks event: %w", err)
	}

	select {
	case msg := <-responseCh:
		logger.Println("Received message from tasks.deleted topic")
		var event map[string]string
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logger.Printf("Invalid confirmation message format: %v", err)
			span.RecordError(err)
			return fmt.Errorf("invalid confirmation message format: %w", err)
		}

		if event["project_id"] != projectID.Hex() {
			err := fmt.Errorf("confirmation received for incorrect project ID")
			logger.Printf(err.Error())
			span.RecordError(err)
			return err
		}

		logger.Printf("Successfully received confirmation for project ID: %s", projectID.Hex())
		_, err = ps.repo.Delete(ctx, projectID)
		if err != nil {
			logger.Printf("Failed to delete project: %v", err)
			span.RecordError(err)
			return fmt.Errorf("failed to delete project: %w", err)
		}

	case <-time.After(20 * time.Second):
		err := fmt.Errorf("timeout waiting for task deletion confirmation")
		logger.Printf("Timeout waiting for task deletion confirmation for project ID: %s", projectID.Hex())
		span.RecordError(err)
		return err
	}

	logger.Printf("Successfully deleted project with ID: %s", projectID.Hex())
	return nil
}

func (ps *ProjectService) GetProjectsByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.Project, error) {
	tracer := otel.Tracer("projects-service")
	ctx, span := tracer.Start(ctx, "GetProjectsByUserIDService")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.Hex()))

	projects, err := ps.repo.GetProjectsByUserID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("projects.count", len(projects)))
	return projects, nil
}

func (ps *ProjectService) GetProjectsByManagerID(ctx context.Context, userID primitive.ObjectID) ([]model.Project, error) {
	tracer := otel.Tracer("projects-service")
	ctx, span := tracer.Start(ctx, "GetProjectsByManagerIDService")
	defer span.End()

	span.SetAttributes(attribute.String("manager.id", userID.Hex()))

	projects, err := ps.repo.GetProjectsByManagerID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("projects.count", len(projects)))
	return projects, nil
}
