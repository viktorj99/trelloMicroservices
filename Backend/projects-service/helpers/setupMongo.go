package helpers

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func SetupMongoClient(ctx context.Context, dburi string) (*mongo.Client, error) {
	tracer := otel.Tracer("projects-service")
	ctx, span := tracer.Start(ctx, "SetupMongoClient")
	defer span.End()

	span.SetAttributes(attribute.String("db.uri", dburi))

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	err = client.Connect(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	return client, nil
}

func ConnectMongoDB(ctx context.Context, logger *log.Logger) *mongo.Client {
	dburi := os.Getenv("MONGO_DB_URI")
	tracer := otel.Tracer("projects-service")
	ctx, span := tracer.Start(ctx, "ConnectMongoDB")
	defer span.End()

	mongoClient, err := SetupMongoClient(ctx, dburi)
	if err != nil {
		span.RecordError(err)
		logger.Fatal("Failed to connect to MongoDB: ", err)
	}
	span.SetAttributes(attribute.String("db.uri", dburi))
	return mongoClient
}
