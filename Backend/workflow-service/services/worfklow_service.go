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

func (service *WorkflowService) CreateTask(task models.Task) error {
	err := service.repo.CreateTask(task)
	if err != nil {
		return err
	}

	// RECOMPUTE BLOCKED FLAG
	blocked, err := service.repo.IsTaskBlocked(task.ID)
	if err != nil {
		return err
	}
	return service.repo.SetTaskBlocked(task.ID, blocked)
}

func (service *WorkflowService) TaskExists(taskID string) (bool, error) {
	return service.repo.TaskExists(taskID)
}

func (service *WorkflowService) CreateDependency(req models.DependencyRequest) error {
	err := service.repo.CreateDependency(req.TaskID, req.DependentID)
	if err != nil {
		return err
	}

	// RECOMPUTE for both involved tasks
	affected := []string{req.TaskID, req.DependentID}
	for _, id := range affected {
		blocked, err := service.repo.IsTaskBlocked(id)
		if err != nil {
			return err
		}
		err = service.repo.SetTaskBlocked(id, blocked)
		if err != nil {
			return err
		}
	}

	return nil
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
	err := service.repo.UpdateTask(task)
	if err != nil {
		return err
	}

	// RECOMPUTE for this task and all its dependents
	dependents, err := service.repo.GetTaskDependents(task.ID)
	if err != nil {
		return err
	}

	allToCheck := append(dependents, task.ID)
	for _, id := range allToCheck {
		blocked, err := service.repo.IsTaskBlocked(id)
		if err != nil {
			return err
		}
		err = service.repo.SetTaskBlocked(id, blocked)
		if err != nil {
			return err
		}
	}

	return nil
}
