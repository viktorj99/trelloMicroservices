package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"projects-service/handlers"
	"projects-service/repositories"
	"projects-service/services"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	// Initialize MongoDB connection
	dburi := os.Getenv("MONGO_DB_URI")
	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		logger.Fatal("Failed to parse URI: ", err)
	}
	err = client.Connect(ctx)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB: ", err)
	}
	defer client.Disconnect(ctx)

	// Initialize repositories, services, and handlers
	repo, err := repositories.NewProjectRepo(ctx, logger, "project")
	if err != nil {
		logger.Fatal("Failed to create repository: ", err)
	}
	service := services.NewProjectService(repo)
	handler := handlers.NewProjectHandler(service)

	// Define routes and start server
	http.HandleFunc("/projects", handler.GetAllProjects)
	http.HandleFunc("/project", handler.GetProjectById)
	http.HandleFunc("/project/create", handler.CreateProject)
	http.HandleFunc("/project/update", handler.UpdateProject)
	http.HandleFunc("/project/delete", handler.DeleteProject)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Printf("Server is starting on port %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.Fatal("Server failed to start: ", err)
	}
}
