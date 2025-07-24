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

	// Initialize logging
	handlers.InitLogging()
	defer handlers.CloseLogging()

	// Initialize OpenTelemetry tracing
	shutdown := helpers.InitTracer()
	defer shutdown()

	ctx := context.Background()
	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	// Connect to MongoDB
	mongoClient := helpers.ConnectMongoDB(ctx, logger)
	defer mongoClient.Disconnect(ctx)

	// Connect to user service
	userClient := client.ConnectToUserService(logger)
	defer userClient.Close()

	// Start the gRPC server for the project service
	helpers.StartGRPCServer()

	// Connect to task service before starting gRPC server
	logger.Println("Connecting to task service...")
	taskServiceClient, err := client.NewTaskClient("tasks-service:50051")
	if err != nil {
		logger.Fatalf("could not connect to task service: %v", err)
	}
	logger.Println("Connected to task service.")
	defer taskServiceClient.Close()

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	natsClient := client.NewNATSClient(natsURL)
	defer natsClient.Conn.Close()

	// Initialize the service and handlers
	service := helpers.InitializeService(ctx, logger, natsURL)
	handler := handlers.NewProjectHandler(service, taskServiceClient, userClient, logger)

	// Setup routes for Project REST API and gRPC
	router := helpers.SetupRoutes(handler, userClient.Client, taskServiceClient)
	helpers.RunServer(router, logger)
}
