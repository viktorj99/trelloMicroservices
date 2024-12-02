package main

import (
	"context"
	"log"
	"os"
	"users-service/helpers"
	"users-service/repositories"
	"users-service/services"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	// err := utils.LoadCommonPasswordsToConsul("/app/config/common_passwords.txt")
	// if err != nil {
	// 	log.Fatalf("Failed to load common passwords to Consul: %v", err)
	// }

	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	dbURI := os.Getenv("MONGO_DB_URI")
	if dbURI == "" {
		logger.Fatal("MONGO_DB_URI is not set")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(dbURI))
	if err != nil {
		logger.Fatal("Failed to parse MongoDB URI: ", err)
	}

	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB: ", err)
	}
	defer client.Disconnect(ctx)

	repositories.InitRepository(client)

	go services.StartKeyExpirationListener()

	// Pokretanje HTTP i gRPC servera paralelno
	go helpers.StartHTTPServer() // Pokrenuti i HTTP i HTTPS
	helpers.StartGRPCServer()    // gRPC server na portu 50051
}
