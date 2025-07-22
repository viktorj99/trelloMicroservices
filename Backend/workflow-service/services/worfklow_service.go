package services

import (
	"context"
	"encoding/json"
	"log"
	"workflow/models"
	"workflow/repositories"

	"github.com/nats-io/nats.go"
)

type WorkflowService struct {
	repo       *repositories.WorkflowRepository
	natsClient *nats.Conn
	logger     *log.Logger
}

func NewWorkflowService(repo *repositories.WorkflowRepository, natsClient *nats.Conn, logger *log.Logger) *WorkflowService {
	return &WorkflowService{
		repo:       repo,
		natsClient: natsClient,
		logger:     logger,
	}
}

func (service *WorkflowService) Start() {
	service.ListenForWorkflowDeletion()
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

func (service *WorkflowService) DeleteWorkflowsForProject(ctx context.Context, projectID string) error {
	service.logger.Printf("🚨 Deleting all workflows for project ID: %s", projectID)

	err := service.repo.DeleteWorkflowsByProjectID(projectID)
	if err != nil {
		service.logger.Printf("❌ Failed to delete workflows for project: %s — %v", projectID, err)
		return err
	}

	// Send confirmation
	confirmation := map[string]string{"project_id": projectID}
	data, err := json.Marshal(confirmation)
	if err != nil {
		service.logger.Printf("❌ Could not marshal deletion confirmation: %v", err)
		return err
	}

	err = service.natsClient.Publish("workflows.deleted", data)
	if err != nil {
		service.logger.Printf("❌ Failed to publish deletion confirmation: %v", err)
		return err
	}

	service.logger.Printf("✅ Workflows deleted and confirmation published for project: %s", projectID)
	return nil
}

func (service *WorkflowService) ListenForWorkflowDeletion() {
	service.logger.Println("📡 Listening on 'workflows.delete' topic...")

	_, err := service.natsClient.Subscribe("workflows.delete", func(msg *nats.Msg) {
		service.logger.Println("📥 Received 'workflows.delete' event")

		var event map[string]string
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			service.logger.Printf("❌ Invalid JSON in delete message: %v", err)
			return
		}

		projectID, ok := event["project_id"]
		if !ok {
			service.logger.Println("⚠️ 'project_id' missing in delete event")
			return
		}

		if err := service.DeleteWorkflowsForProject(context.Background(), projectID); err != nil {
			service.logger.Printf("❌ Error while deleting workflows for project %s: %v", projectID, err)
		}
	})

	if err != nil {
		service.logger.Fatalf("❌ Failed to subscribe to 'workflows.delete': %v", err)
	}
}
