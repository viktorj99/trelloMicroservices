package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	taskpb "pb/taskpb"
	"strings"
	"tasks-service/handlers"
	"tasks-service/repositories"
	"tasks-service/server"
	"tasks-service/services"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
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

	router := mux.NewRouter()

	router.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/tasks/") && len(r.URL.Path) > len("/tasks/") {
			taskHandler.GetTaskById(w, r)
		} else {
			taskHandler.GetAllTasks(w, r)
		}
	})

	router.HandleFunc("/tasks", taskHandler.GetAllTasks).Methods("GET")
	router.HandleFunc("/tasks/{id}", taskHandler.GetTaskById).Methods("GET")
	router.HandleFunc("/tasks/{projectId}/tasks", taskHandler.GetTasksByProjectId).Methods("GET")
	router.HandleFunc("/tasks/create", taskHandler.CreateTask).Methods("POST")
	router.HandleFunc("/tasks/{taskID}/assign/{memberID}", taskHandler.AssignMemberToTask).Methods("PUT")
	router.HandleFunc("/tasks/{taskID}/member/{memberID}/toggle-status", taskHandler.ToggleTaskStatus).Methods("PUT")
	router.HandleFunc("/tasks/{taskID}/remove-member", taskHandler.RemoveMemberFromTask).Methods("PUT")

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080" // Default HTTP port
	}

	httpsPort := os.Getenv("HTTPS_PORT")
	if httpsPort == "" {
		httpsPort = "8443" // Default HTTPS port
	}

	// Channels to catch server errors
	errChan := make(chan error)

	// Start HTTP server
	go func() {
		logger.Printf("HTTP server is starting on port %s...", httpPort)
		errChan <- http.ListenAndServe(":"+httpPort, nil)
	}()

	// Start HTTPS server
	go func() {
		logger.Printf("HTTPS server is starting on port %s...", httpsPort)
		errChan <- http.ListenAndServeTLS(":"+httpsPort, "certificates/cert.crt", "certificates/cert.key", router)
	}()

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()
		taskServer := server.NewTaskServer(taskRepo)
		taskpb.RegisterTaskServiceServer(grpcServer, taskServer)

		logger.Println("gRPC server is running on port 50051...")
		errChan <- grpcServer.Serve(lis)
	}()

	// Wait for an error to occur
	err = <-errChan
	log.Fatalf("Server error: %v", err)
}
