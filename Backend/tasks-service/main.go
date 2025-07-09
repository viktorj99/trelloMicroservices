package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	taskpb "pb/taskpb"
	"strings"
	"tasks-service/client"
	"tasks-service/handlers"
	"tasks-service/helpers"
	"tasks-service/repositories"
	"tasks-service/server"
	"tasks-service/services"

	"github.com/colinmarc/hdfs"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
)

func main() {
	cleanup := helpers.InitTracer()
	defer cleanup()

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found or failed to load")
	}

	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	ctx, span := otel.Tracer("tasks-service").Start(context.Background(), "main")
	defer span.End()

	dbURI := os.Getenv("MONGO_DB_URI")
	if dbURI == "" {
		logger.Fatal("MONGO_DB_URI is not set")
	}
	span.SetAttributes(attribute.String("db.uri", dbURI))

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(dbURI))
	if err != nil {
		span.RecordError(err)
		logger.Fatal("Failed to parse MongoDB URI: ", err)
	}

	err = mongoClient.Connect(ctx)
	if err != nil {
		span.RecordError(err)
		logger.Fatal("Failed to connect to MongoDB: ", err)
	}
	defer mongoClient.Disconnect(ctx)

	taskRepo, err := repositories.NewTaskRepo(ctx, logger, "task")
	if err != nil {
		span.RecordError(err)
		logger.Fatal("Failed to create repository: ", err)
	}

	documentRepo, err := repositories.NewDocumentRepo(ctx, logger, "documents")
	if err != nil {
		span.RecordError(err)
		logger.Fatal("Failed to create document repository: ", err)
	}

	hdfsClient, err := hdfs.New("namenode.hadoop:9000")
	if err != nil {
		span.RecordError(err)
		logger.Fatal("Failed to create HDFS client: ", err)
	}

	documentService := services.NewDocumentService(documentRepo, hdfsClient)

	errChan := make(chan error)

	go func() {
		_, grpcSpan := otel.Tracer("tasks-service").Start(ctx, "Start gRPC Server")
		defer grpcSpan.End()

		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			grpcSpan.RecordError(err)
			log.Fatalf("failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()
		taskServer := server.NewTaskServer(taskRepo)
		taskpb.RegisterTaskServiceServer(grpcServer, taskServer)

		logger.Println("gRPC server is running on port 50051...")
		if err := grpcServer.Serve(lis); err != nil {
			grpcSpan.RecordError(err)
			log.Fatalf("failed to serve gRPC server: %v", err)
		}
	}()

	logger.Println("Connecting to project service...")
	projectServiceClient, err := client.NewProjectClient("projects-service:50051")
	if err != nil {
		span.RecordError(err)
		logger.Fatalf("could not connect to project service: %v", err)
	}
	logger.Println("Connected to project service.")
	defer projectServiceClient.Close()

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}
	natsClient := client.NewNATSClient(natsURL)
	defer natsClient.Conn.Close()

	taskService := services.NewTaskService(taskRepo, natsClient, logger)
	go taskService.Start()

	taskHandler := handlers.NewTaskHandler(taskService, logger, projectServiceClient)
	documentHandler := handlers.NewDocumentHandler(documentService, logger) // Create Document Handler
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

	router.HandleFunc("/tasks/{taskId}/documents", documentHandler.UploadDocumentHandler).Methods("POST")
	router.HandleFunc("/tasks/{taskId}/documents", documentHandler.GetDocumentsHandler).Methods("GET")
	router.HandleFunc("/tasks/{taskId}/documents/{docName}/download", documentHandler.DownloadDocumentHandler).Methods("GET")
	router.HandleFunc("/tasks/{taskId}/documents/{docName}", documentHandler.DeleteDocumentHandler).Methods("DELETE")

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	go func() {
		_, httpSpan := otel.Tracer("tasks-service").Start(ctx, "Start HTTP Server")
		defer httpSpan.End()

		logger.Printf("HTTP server is starting on port %s...", httpPort)
		errChan <- http.ListenAndServe(":"+httpPort, nil)
	}()

	httpsPort := os.Getenv("HTTPS_PORT")
	if httpsPort == "" {
		httpsPort = "8443"
	}
	go func() {
		_, httpsSpan := otel.Tracer("tasks-service").Start(ctx, "Start HTTPS Server")
		defer httpsSpan.End()

		logger.Printf("HTTPS server is starting on port %s...", httpsPort)
		errChan <- http.ListenAndServeTLS(":"+httpsPort, "certificates/cert.crt", "certificates/cert.key", router)
	}()

	err = <-errChan
	span.RecordError(err)
	log.Fatalf("Server error: %v", err)
}
