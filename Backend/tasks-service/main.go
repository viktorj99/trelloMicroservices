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
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found or failed to load")
	}

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

	taskRepo, err := repositories.NewTaskRepo(ctx, logger, "task")
	if err != nil {
		logger.Fatal("Failed to create repository: ", err)
	}
	taskService := services.NewTaskService(taskRepo, logger)
	taskHandler := handlers.NewTaskHandler(taskService, logger)

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/tasks/") && len(r.URL.Path) > len("/tasks/") {
			taskHandler.GetTaskById(w, r)
		} else {
			taskHandler.GetAllTasks(w, r)
		}
	})

	http.HandleFunc("/tasks/create", taskHandler.CreateTask)

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080" // Default HTTP port
	}

	httpsPort := os.Getenv("HTTPS_PORT")
	if httpsPort == "" {
		httpsPort = "8443" // Default HTTPS port
	}

	go func() {
		logger.Printf("HTTP server is starting on port %s...", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			logger.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	logger.Printf("HTTPS server is starting on port %s...", httpsPort)
	if err := http.ListenAndServeTLS(":"+httpsPort, "certificates/cert.crt", "certificates/cert.key", nil); err != nil {
		logger.Fatalf("HTTPS server failed to start: %v", err)
	}
}
