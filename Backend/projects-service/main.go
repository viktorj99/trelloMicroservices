package main

import (
	"context"
	"log"
	"os"
	"projects-service/client"

	"projects-service/handlers"
	"projects-service/helpers"
)

func main() {
	// helpers.LoadingEnv()

	ctx := context.Background()
	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	// Povezivanje sa MongoDB
	mongoClient := helpers.ConnectMongoDB(ctx, logger)
	defer mongoClient.Disconnect(ctx)

	// Povezivanje sa gRPC user-service preko novog klijenta
	userClient := client.ConnectToUserService(logger)
	defer userClient.Close()

	logger.Println("Connecting to task service...")
	taskServiceClient, err := client.NewTaskClient("tasks-service:50051")
	if err != nil {
		logger.Fatalf("could not connect to task service: %v", err)
	}
	logger.Println("Connected to task service.")
	defer taskServiceClient.Close()

	// Inicijalizacija servisa
	service := helpers.InitializeService(ctx, logger)
	handler := handlers.NewProjectHandler(service, taskServiceClient)

	// Postavljanje ruta za Project REST API i GRPC
	router := helpers.SetupRoutes(handler, userClient.Client, taskServiceClient)
	helpers.RunServer(router, logger)

}
