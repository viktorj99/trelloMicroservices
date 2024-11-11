package helpers

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetupMongoClient(ctx context.Context, dburi string) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func ConnectMongoDB(ctx context.Context, logger *log.Logger) *mongo.Client {
	dburi := os.Getenv("MONGO_DB_URI")
	mongoClient, err := SetupMongoClient(ctx, dburi)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB: ", err)
	}
	return mongoClient
}