package repositories

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tasks-service/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// DocumentRepo handles database operations for documents
type DocumentRepo struct {
	cli    *mongo.Client
	logger *log.Logger
	dbName string
}

// NewDocumentRepo creates a new instance of DocumentRepo
func NewDocumentRepo(ctx context.Context, logger *log.Logger, dbName string) (*DocumentRepo, error) {
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

	return &DocumentRepo{
		cli:    client,
		logger: logger,
		dbName: dbName,
	}, nil
}

// Close disconnects the MongoDB client
func (dr *DocumentRepo) Close(ctx context.Context) error {
	return dr.cli.Disconnect(ctx)
}

// collection returns the MongoDB collection for documents
func (dr *DocumentRepo) collection() *mongo.Collection {
	return dr.cli.Database(dr.dbName).Collection("documents")
}

// InsertDocument inserts a document into the database
func (dr *DocumentRepo) InsertDocument(ctx context.Context, doc *model.Document) (*mongo.InsertOneResult, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "InsertDocumentRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "insert"),
		attribute.String("db.collection", "documents"),
		attribute.String("document.task_id", doc.TaskID.Hex()),
	)

	result, err := dr.collection().InsertOne(ctx, doc)
	if err != nil {
		dr.logger.Println("Error inserting document:", err)
		span.RecordError(err)
		return nil, fmt.Errorf("failed to insert document: %v", err)
	}

	span.SetAttributes(attribute.String("db.inserted_id", result.InsertedID.(primitive.ObjectID).Hex()))
	return result, nil
}

// FindDocumentsByTaskID finds documents by task ID
func (dr *DocumentRepo) FindDocumentsByTaskID(ctx context.Context, taskID string) ([]model.Document, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "FindDocumentsByTaskIDRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "find"),
		attribute.String("db.collection", "documents"),
		attribute.String("task.id", taskID),
	)

	// ✅ Convert string to ObjectID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		dr.logger.Println("Invalid task ID:", err)
		span.RecordError(err)
		return nil, fmt.Errorf("invalid task ID: %v", err)
	}

	var documents []model.Document
	cursor, err := dr.collection().Find(ctx, bson.M{"task_id": objID})
	if err != nil {
		dr.logger.Println("Error retrieving documents by task ID:", err)
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find documents: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc model.Document
		if err := cursor.Decode(&doc); err != nil {
			dr.logger.Println("Error decoding document:", err)
			span.RecordError(err)
			return nil, fmt.Errorf("failed to decode document: %v", err)
		}
		documents = append(documents, doc)
	}

	if err := cursor.Err(); err != nil {
		dr.logger.Println("Cursor error:", err)
		span.RecordError(err)
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	span.SetAttributes(attribute.Int("documents.count", len(documents)))
	return documents, nil
}

// FindDocumentByID finds a single document by its ID
func (dr *DocumentRepo) FindDocumentByID(ctx context.Context, id string) (*model.Document, error) {
	tracer := otel.Tracer("tasks-service/repository")
	ctx, span := tracer.Start(ctx, "FindDocumentByIDRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "findOne"),
		attribute.String("db.collection", "documents"),
		attribute.String("document.id", id),
	)

	var doc model.Document
	err := dr.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		dr.logger.Println("Error retrieving document by ID:", err)
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find document: %v", err)
	}

	return &doc, nil
}

func (dr *DocumentRepo) DeleteDocument(filter bson.M) (*mongo.DeleteResult, error) {
	return dr.collection().DeleteOne(context.Background(), filter)
}

func (dr *DocumentRepo) DeleteDocumentByTaskIDAndName(ctx context.Context, taskID, fileName string) error {
	objectID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return fmt.Errorf("invalid task ID: %v", err)
	}

	filter := bson.M{
		"task_id":   objectID,
		"file_name": fileName,
	}

	result, err := dr.collection().DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete document: %v", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no document found to delete")
	}

	return nil
}
