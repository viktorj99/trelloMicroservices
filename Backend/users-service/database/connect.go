package database

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientInstance *mongo.Client
var clientInstanceError error
var mongoOnce sync.Once

func GetMongoClient() (*mongo.Client, error) {
	mongoOnce.Do(func() {
		mongoURI := os.Getenv("MONGO_DB_URI")
		if mongoURI == "" {
			log.Fatal("MONGO_DB_URI is not set in the environment variables")
		}

		clientOptions := options.Client().ApplyURI(mongoURI)
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			clientInstanceError = err
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = client.Ping(ctx, nil)
		if err != nil {
			clientInstanceError = err
			return
		}

		log.Println("Connected to MongoDB!")
		clientInstance = client
	})

	return clientInstance, clientInstanceError
}
