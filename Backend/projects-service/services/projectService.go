package services

import (
	"context"
	"errors"
	"projects-service/model"
	"projects-service/repositories"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProjectService struct {
	repo *repositories.ProjectRepo
}

func NewProjectService(repo *repositories.ProjectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

func (ps *ProjectService) GetAllProjects(ctx context.Context) ([]model.Project, error) {
	return ps.repo.GetAll(ctx)
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

func (ps *ProjectService) DeleteProject(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	if id.IsZero() {
		return nil, errors.New("invalid project ID")
	}

	return ps.repo.Delete(ctx, id)
}
