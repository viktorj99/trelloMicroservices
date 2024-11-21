package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"tasks-service/handlers"
	"tasks-service/repositories"
	"tasks-service/services"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
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

	// Adjusted route handling for consistency
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/tasks/") && len(r.URL.Path) > len("/tasks/") {
			taskHandler.GetTaskById(w, r)
		} else {
			taskHandler.GetAllTasks(w, r)
		}
	})

	http.HandleFunc("/tasks/create", taskHandler.CreateTask)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	logger.Printf("Server is starting on port %s", port)
	err = http.ListenAndServe(":"+port, http.DefaultServeMux)
	if err != nil {
		logger.Fatal("Server failed to start: ", err)
	}
}
