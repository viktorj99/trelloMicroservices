package services

import (
	"context"
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
	return ps.repo.GetById(ctx, id)
}

func (ps *ProjectService) CreateProject(ctx context.Context, project *model.Project) (*mongo.InsertOneResult, error) {
	return ps.repo.Insert(ctx, project)
}

func (ps *ProjectService) UpdateProject(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*mongo.UpdateResult, error) {
	return ps.repo.Update(ctx, id, updateData)
}

func (ps *ProjectService) DeleteProject(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return ps.repo.Delete(ctx, id)
}
