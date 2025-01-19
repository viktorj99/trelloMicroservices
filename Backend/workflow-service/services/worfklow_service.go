package services

import (
	"fmt"
	"workflow/models"
	"workflow/repositories"
)

type WorkflowService struct {
	repo *repositories.WorkflowRepository
}

func NewWorkflowService(repo *repositories.WorkflowRepository) *WorkflowService {
	return &WorkflowService{repo: repo}
}

func (service *WorkflowService) CreateTask(task models.Task) error {
	return service.repo.CreateTask(task)
}

func (service *WorkflowService) TaskExists(taskID string) (bool, error) {
	return service.repo.TaskExists(taskID)
}

func (service *WorkflowService) CreateDependency(req models.DependencyRequest) error {
	return service.repo.CreateDependency(req.TaskID, req.DependentID)
}

func (service *WorkflowService) GetTasks() ([]map[string]interface{}, error) {
	return service.repo.GetTasks()
}

func (service *WorkflowService) GetTasksWithDependencies(projectID string) ([]map[string]interface{}, error) {
	return service.repo.GetTasksWithDependencies(projectID)
}

func (service *WorkflowService) GetAllDependencies() ([]map[string]interface{}, error) {
	return service.repo.GetAllDependencies()
}

func (service *WorkflowService) GetTask(taskID string) (models.Task, error) {
	return service.repo.GetTask(taskID)
}

func (service *WorkflowService) UpdateTask(task models.Task) error {
	fmt.Println("Pozvao servis za update task ssssss", task.Status)
	return service.repo.UpdateTask(task)
}
