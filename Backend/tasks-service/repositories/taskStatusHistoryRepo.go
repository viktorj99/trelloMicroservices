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

type TaskStatusHistoryRepo struct {
	cli    *mongo.Client
	logger *log.Logger
	dbName string
}

func NewTaskStatusHistoryRepo(ctx context.Context, logger *log.Logger, dbName string) (*TaskStatusHistoryRepo, error) {
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

	return &TaskStatusHistoryRepo{
		cli:    client,
		logger: logger,
		dbName: dbName,
	}, nil
}

func (tr *TaskStatusHistoryRepo) Close(ctx context.Context) error {
	return tr.cli.Disconnect(ctx)
}

func (tr *TaskStatusHistoryRepo) collection() *mongo.Collection {
	return tr.cli.Database(tr.dbName).Collection("task_status_history")
}

func (tr *TaskStatusHistoryRepo) Insert(ctx context.Context, history *model.TaskStatusHistory) (*mongo.InsertOneResult, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "InsertTaskStatusHistoryRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "insert"),
		attribute.String("db.collection", "task_status_history"),
		attribute.String("task.id", history.TaskID.Hex()),
		attribute.String("status", string(history.Status)),
	)

	history.ID = primitive.NewObjectID()
	history.Timestamp = time.Now()

	result, err := tr.collection().InsertOne(ctx, history)
	if err != nil {
		tr.logger.Println("Error inserting task status history:", err)
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.String("db.inserted_id", result.InsertedID.(primitive.ObjectID).Hex()))
	return result, nil
}

func (tr *TaskStatusHistoryRepo) GetByTaskId(ctx context.Context, taskId primitive.ObjectID) ([]model.TaskStatusHistory, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "GetTaskStatusHistoryByTaskIdRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "find"),
		attribute.String("db.collection", "task_status_history"),
		attribute.String("task.id", taskId.Hex()),
	)

	var histories []model.TaskStatusHistory
	cursor, err := tr.collection().Find(ctx, bson.M{"taskId": taskId})
	if err != nil {
		tr.logger.Println("Error retrieving task status history:", err)
		span.RecordError(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &histories); err != nil {
		tr.logger.Println("Error decoding task status history:", err)
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("history.count", len(histories)))
	return histories, nil
}
