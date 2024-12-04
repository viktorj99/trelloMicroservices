package services

import (
	"workflow/models"
	"workflow/repositories"
)

type WorkflowService struct {
	repo *repositories.WorkflowRepository
}

func NewWorkflowService(repo *repositories.WorkflowRepository) *WorkflowService {
	return &WorkflowService{repo: repo}
}

func (s *WorkflowService) CreateTask(task *models.Task) error {
	return s.repo.CreateTaskNode(task)
}

// CreateDependency delegates dependency creation to the repository
func (s *WorkflowService) CreateDependency(dep *models.Dependency) error {
	return s.repo.CreateDependency(dep)
}
