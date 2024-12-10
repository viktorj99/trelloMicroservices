package repositories

import (
	"fmt"
	"workflow/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type WorkflowRepository struct {
	driver neo4j.Driver
}

func NewWorkflowRepository(driver neo4j.Driver) *WorkflowRepository {
	return &WorkflowRepository{driver: driver}
}

func (repo *WorkflowRepository) CreateTask(task models.Task) error {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	_, err := session.Run(
		`CREATE (t:Task {
			id: $id, 
			title: $title, 
			description: $description,
			status: $status,
			project: $project,
			member: $member,
			isBlocked: $isBlocked
		})`,
		map[string]interface{}{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"project":     task.Project,
			"member":      task.Member,
			"isBlocked":   task.IsBlocked,
		},
	)
	return err
}

func (repo *WorkflowRepository) TaskExists(taskID string) (bool, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	query := `
		MATCH (t:Task {id: $id})
		RETURN COUNT(t) > 0 AS exists
	`

	result, err := session.Run(query, map[string]interface{}{
		"id": taskID,
	})
	if err != nil {
		return false, err
	}

	if result.Next() {
		return result.Record().Values[0].(bool), nil
	}

	return false, nil
}


func (repo *WorkflowRepository) CreateDependency(taskID, dependentID string) error {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	// Provera ciklusa
	result, err := session.Run(
		"MATCH (t1:Task {id: $taskID}), (t2:Task {id: $dependentID}) "+
			"CALL apoc.algo.dijkstra(t2, t1, 'DEPENDS_ON', 'weight') YIELD path "+
			"RETURN path",
		map[string]interface{}{
			"taskID":      taskID,
			"dependentID": dependentID,
		},
	)
	if err != nil {
		return err
	}
	if result.Next() {
		return fmt.Errorf("dependency creates a cycle")
	}

	_, err = session.Run(
		"MATCH (t1:Task {id: $taskID}), (t2:Task {id: $dependentID}) "+
			"CREATE (t1)-[:DEPENDS_ON]->(t2)",
		map[string]interface{}{
			"taskID":      taskID,
			"dependentID": dependentID,
		},
	)
	return err
}


func (repo *WorkflowRepository) GetTasks() ([]map[string]interface{}, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	query := `
		MATCH (t:Task)
		RETURN t
	`

	result, err := session.Run(query, nil)
	if err != nil {
		return nil, err
	}

	var tasks []map[string]interface{}
	for result.Next() {
		record := result.Record()
		taskNode := record.Values[0].(neo4j.Node)
		tProps := taskNode.Props

		task := map[string]interface{}{
			"id":          tProps["id"],
			"title":       tProps["title"],
			"description": tProps["description"],
			"status":      tProps["status"],
			"project":     tProps["project"],
			"member":      tProps["member"],
			"blocked":     tProps["isBlocked"],
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (repo *WorkflowRepository) GetTasksWithDependencies() ([]map[string]interface{}, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	query := `
		MATCH (t:Task)
		OPTIONAL MATCH (t)-[:DEPENDS_ON]->(dependent:Task)
		RETURN t, collect(dependent) AS dependencies
	`

	result, err := session.Run(query, nil)
	if err != nil {
		return nil, err
	}

	var tasks []map[string]interface{}
	for result.Next() {
		record := result.Record()
		taskNode := record.Values[0].(neo4j.Node)
		tProps := taskNode.Props

		task := map[string]interface{}{
			"id":          tProps["id"],
			"title":       tProps["title"],
			"description": tProps["description"],
			"status":      tProps["status"],
			"project":     tProps["project"],
			"member":      tProps["member"],
			"blocked":     tProps["isBlocked"],
		}

		// Dependencies handling
		depsRaw := record.Values[1].([]interface{})
		var deps []map[string]interface{}
		for _, d := range depsRaw {
			if d != nil {
				depNode := d.(neo4j.Node)
				depProps := depNode.Props
				deps = append(deps, map[string]interface{}{
					"id":          depProps["id"],
					"title":       depProps["title"],
					"description": depProps["description"],
					"status":      depProps["status"],
					"project":     depProps["project"],
					"member":      depProps["member"],
					"blocked":     depProps["isBlocked"],
				})
			}
		}
		task["dependencies"] = deps

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (repo *WorkflowRepository) GetAllDependencies() ([]map[string]interface{}, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	query := `
		MATCH (t1:Task)-[r:DEPENDS_ON]->(t2:Task)
		RETURN t1, t2
	`

	result, err := session.Run(query, nil)
	if err != nil {
		return nil, err
	}

	var dependencies []map[string]interface{}
	for result.Next() {
		record := result.Record()

		task1 := record.Values[0].(neo4j.Node).Props
		task2 := record.Values[1].(neo4j.Node).Props

		dependency := map[string]interface{}{
			"task":         task1,
			"dependentOn":  task2,
		}
		dependencies = append(dependencies, dependency)
	}

	return dependencies, nil
}
