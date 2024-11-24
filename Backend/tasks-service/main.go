package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"tasks-service/handlers"
	"tasks-service/repositories"
	"tasks-service/services"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	dbURI := os.Getenv("MONGO_DB_URI")
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

	taskRepo, err := repositories.NewTaskRepo(ctx, logger, "task")
	if err != nil {
		logger.Fatal("Failed to create repository: ", err)
	}
	taskService := services.NewTaskService(taskRepo, logger)
	taskHandler := handlers.NewTaskHandler(taskService, logger)

	router := mux.NewRouter()

	router.HandleFunc("/tasks", taskHandler.GetAllTasks).Methods("GET")
	router.HandleFunc("/tasks/{id}", taskHandler.GetTaskById).Methods("GET")
	router.HandleFunc("/tasks/{projectId}/tasks", taskHandler.GetTasksByProjectId).Methods("GET")
	router.HandleFunc("/tasks/create", taskHandler.CreateTask).Methods("POST")
	router.HandleFunc("/tasks/{taskID}/assign/{memberID}", taskHandler.AssignMemberToTask).Methods("PUT")
	router.HandleFunc("/tasks/{taskID}/member/{memberID}/toggle-status", taskHandler.ToggleTaskStatus).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	logger.Printf("Server is starting on port %s", port)

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		logger.Fatal("Server failed to start: ", err)
	}
}
