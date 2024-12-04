package repositories

import (
	"workflow/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type WorkflowRepository struct {
	Driver neo4j.Driver
}

func NewWorkflowRepository(driver neo4j.Driver) *WorkflowRepository {
	return &WorkflowRepository{Driver: driver}
}

func (r *WorkflowRepository) CreateTaskNode(task *models.Task) error {
	session := r.Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	_, err := session.Run(`
		CREATE (t:Task {
			id: $id,
			projectID: $projectID,
			name: $name,
			description: $description,
			blocked: $blocked
		})
	`, map[string]interface{}{
		"id":          task.ID,
		"projectID":   task.ProjectID,
		"name":        task.Name,
		"description": task.Description,
		"blocked":     task.Blocked,
	})
	return err
}

func (r *WorkflowRepository) CreateDependency(dep *models.Dependency) error {
	session := r.Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	_, err := session.Run(`
		MATCH (a:Task {id: $fromTaskID}), (b:Task {id: $toTaskID})
		CREATE (a)-[:DEPENDS_ON]->(b)
		SET b.blocked = true
	`, map[string]interface{}{
		"fromTaskID": dep.FromTaskID,
		"toTaskID":   dep.ToTaskID,
	})
	return err
}
