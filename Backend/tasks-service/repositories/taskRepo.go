package repositories

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tasks-service/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "GetAllTasksRepo")
	defer span.End()

	span.SetAttributes(attribute.String("db.operation", "find"), attribute.String("db.collection", "tasks"))

	var tasks []model.Task
	cursor, err := tr.collection().Find(ctx, bson.M{})
	if err != nil {
		tr.logger.Println("Error retrieving all tasks:", err)
		span.RecordError(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &tasks); err != nil {
		tr.logger.Println("Error decoding all tasks:", err)
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))
	return tasks, nil
}

func (tr *TaskRepo) GetById(ctx context.Context, id primitive.ObjectID) (*model.Task, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "GetTaskByIdRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "findOne"),
		attribute.String("db.collection", "tasks"),
		attribute.String("task.id", id.Hex()),
	)

	var task model.Task
	err := tr.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&task)
	if err != nil {
		tr.logger.Println("Error retrieving task by ID:", err)
		span.RecordError(err)
		if err == mongo.ErrNoDocuments {
			span.SetAttributes(attribute.Bool("db.result.empty", true))
		}
		return nil, err
	}

	return &task, nil
}

func (tr *TaskRepo) Insert(ctx context.Context, task *model.Task) (*mongo.InsertOneResult, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "InsertTaskRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "insert"),
		attribute.String("db.collection", "tasks"),
		attribute.String("task.title", task.Title),
	)

	task.ID = primitive.NewObjectID()
	result, err := tr.collection().InsertOne(ctx, task)
	if err != nil {
		tr.logger.Println("Error inserting task:", err)
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.String("db.inserted_id", result.InsertedID.(primitive.ObjectID).Hex()))
	return result, nil
}

func (tr *TaskRepo) GetByProjectId(ctx context.Context, projectId primitive.ObjectID) ([]model.Task, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "GetTasksByProjectIdRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "find"),
		attribute.String("db.collection", "tasks"),
		attribute.String("project.id", projectId.Hex()),
	)

	var tasks []model.Task
	filter := bson.M{"project": projectId}
	cursor, err := tr.collection().Find(ctx, filter)
	if err != nil {
		tr.logger.Println("Error retrieving tasks by project ID:", err)
		span.RecordError(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &tasks); err != nil {
		tr.logger.Println("Error decoding tasks by project ID:", err)
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))
	return tasks, nil
}

func (tr *TaskRepo) Update(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*mongo.UpdateResult, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "UpdateTaskRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "updateOne"),
		attribute.String("db.collection", "tasks"),
		attribute.String("task.id", id.Hex()),
	)

	result, err := tr.collection().UpdateOne(ctx, bson.M{"_id": id}, updateData)
	if err != nil {
		tr.logger.Println("Error updating task:", err)
		span.RecordError(err)
		return nil, err
	}

	return result, nil
}

func (tr *TaskRepo) GetUnassignedTasks(ctx context.Context, projectId primitive.ObjectID) ([]model.Task, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "GetUnassignedTasksRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "find"),
		attribute.String("db.collection", "tasks"),
		attribute.String("project.id", projectId.Hex()),
	)

	var tasks []model.Task
	filter := bson.M{
		"project": projectId,
		"$or": []bson.M{
			{"member": bson.M{"$exists": false}},
			{"member": primitive.NilObjectID},
			{"member": primitive.NewObjectIDFromTimestamp(time.Time{})},
		},
	}
	cursor, err := tr.collection().Find(ctx, filter)
	if err != nil {
		tr.logger.Println("Error retrieving unassigned tasks:", err)
		span.RecordError(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &tasks); err != nil {
		tr.logger.Println("Error decoding unassigned tasks:", err)
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))
	return tasks, nil
}

func (tr *TaskRepo) HasPendingOrInProgressTasks(ctx context.Context, memberId primitive.ObjectID) (bool, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "CheckPendingOrInProgressTasksRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "countDocuments"),
		attribute.String("db.collection", "tasks"),
		attribute.String("member.id", memberId.Hex()),
	)

	filter := bson.M{
		"member": memberId,
		"status": bson.M{
			"$in": []model.Status{model.InProgress, model.Pending},
		},
	}

	count, err := tr.collection().CountDocuments(ctx, filter)
	if err != nil {
		tr.logger.Println("Error checking pending or in-progress tasks for member:", err)
		span.RecordError(err)
		return false, err
	}

	span.SetAttributes(attribute.Int64("tasks.count", count))
	return count > 0, nil
}

func (tr *TaskRepo) DeleteTasks(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "DeleteTaskRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "deleteOne"),
		attribute.String("db.collection", "tasks"),
		attribute.String("task.id", id.Hex()),
	)

	result, err := tr.collection().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		tr.logger.Println("Error deleting task:", err)
		span.RecordError(err)
		return nil, err
	}

	return result, nil
}
