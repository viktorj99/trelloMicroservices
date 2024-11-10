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

func enableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

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

	router := mux.NewRouter()
	router.HandleFunc("/tasks", taskHandler.GetAllTasks).Methods("GET")
	router.HandleFunc("/task/{id}", taskHandler.GetTaskById).Methods("GET")
	router.HandleFunc("/task/create", taskHandler.CreateTask).Methods("POST")

	http.Handle("/", enableCORS(router))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	logger.Printf("Server is starting on port %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.Fatal("Server failed to start: ", err)
	}
}
