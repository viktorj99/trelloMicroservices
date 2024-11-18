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

	// Inicijalizacija servisa
	service := helpers.InitializeService(ctx, logger)
	handler := handlers.NewProjectHandler(service)

	// Postavljanje ruta za Project REST API i GRPC
	router := helpers.SetupRoutes(handler, userClient.Client)
	helpers.RunServer(router, logger)
}
