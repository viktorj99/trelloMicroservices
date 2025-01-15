package repositories

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tasks-service/model"
)

type TaskRepo struct {
	cli    *mongo.Client
	logger *log.Logger
	dbName string
}

func NewTaskRepo(ctx context.Context, logger *log.Logger, dbName string) (*TaskRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")
	if dburi == "" {
		return nil, fmt.Errorf("environment variable MONGO_DB_URI is not set")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		logger.Println("Failed to create MongoDB client:", err)
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		logger.Println("Failed to connect to MongoDB:", err)
		return nil, err
	}

	return &TaskRepo{
		cli:    client,
		logger: logger,
		dbName: dbName,
	}, nil
}

func (tr *TaskRepo) Close(ctx context.Context) error {
	return tr.cli.Disconnect(ctx)
}

func (tr *TaskRepo) collection() *mongo.Collection {
	return tr.cli.Database(tr.dbName).Collection("tasks")
}

func (tr *TaskRepo) GetAll(ctx context.Context) ([]model.Task, error) {
	var tasks []model.Task
	cursor, err := tr.collection().Find(ctx, bson.M{})
	if err != nil {
		tr.logger.Println("Error retrieving all tasks:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &tasks); err != nil {
		tr.logger.Println("Error decoding all tasks:", err)
		return nil, err
	}

	return tasks, nil
}

func (tr *TaskRepo) GetById(ctx context.Context, id primitive.ObjectID) (*model.Task, error) {
	var task model.Task
	err := tr.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&task)
	if err != nil {
		tr.logger.Println("Error retrieving task by ID:", err)
		return nil, err
	}

	return &task, nil
}

func (tr *TaskRepo) Insert(ctx context.Context, task *model.Task) (*mongo.InsertOneResult, error) {
	task.ID = primitive.NewObjectID()
	result, err := tr.collection().InsertOne(ctx, task)
	if err != nil {
		tr.logger.Println("Error inserting task:", err)
		return nil, err
	}

	return result, nil
}

func (tr *TaskRepo) GetByProjectId(ctx context.Context, projectId primitive.ObjectID) ([]model.Task, error) {
	var tasks []model.Task

	filter := bson.M{"project": projectId}

	cursor, err := tr.collection().Find(ctx, filter)
	if err != nil {
		tr.logger.Println("Error retrieving tasks by project ID:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &tasks); err != nil {
		tr.logger.Println("Error decoding tasks by project ID:", err)
		return nil, err
	}

	return tasks, nil
}

func (tr *TaskRepo) Update(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*mongo.UpdateResult, error) {
	filter := bson.M{"_id": id}

	result, err := tr.collection().UpdateOne(ctx, filter, updateData)
	if err != nil {
		tr.logger.Println("Error updating task:", err)
		return nil, err
	}

	return result, nil
}

func (tr *TaskRepo) GetTasksWithDependencies(ctx context.Context, projectId string) ([]model.TaskWithDependencies, error) {
	// Convert project ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		tr.logger.Println("Invalid project ID format:", err)
		return nil, err
	}

	// Fetch tasks for the project
	tasks, err := tr.GetByProjectId(ctx, objectID)
	if err != nil {
		tr.logger.Println("Error getting tasks:", err)
		return nil, err
	}

	// Call workflow service to get task dependencies
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := httpClient.Get("https://workflow-service:8443/workflow/tasks/dependencies/" + projectId)
	if err != nil {
		tr.logger.Println("Error calling workflow service:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tr.logger.Printf("Workflow service returned non-200 status: %d", resp.StatusCode)
		return nil, fmt.Errorf("workflow service error: %d", resp.StatusCode)
	}

	// Decode workflow response
	var workflowTasks []model.TaskWithDependencies
	if err := json.NewDecoder(resp.Body).Decode(&workflowTasks); err != nil {
		tr.logger.Println("Error decoding workflow response:", err)
		return nil, err
	}

	// Create a map for quick lookup of dependencies by task ID
	workflowTaskMap := make(map[primitive.ObjectID][]model.Task)
	for _, wfTask := range workflowTasks {
		workflowTaskMap[wfTask.ID] = wfTask.Dependencies
	}

	// Combine tasks with dependencies
	tasksWithDeps := make([]model.TaskWithDependencies, len(tasks))
	for i, task := range tasks {
		dependencies := workflowTaskMap[task.ID]
		tasksWithDeps[i] = model.TaskWithDependencies{
			Task:         task,
			Dependencies: dependencies,
		}
	}

	return tasksWithDeps, nil
}


