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

func enableCORS(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handlerFunc(w, r)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

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

	repo, err := repositories.NewProjectRepo(ctx, logger, "project")
	if err != nil {
		logger.Fatal("Failed to create repository: ", err)
	}
	service := services.NewProjectService(repo)
	handler := handlers.NewProjectHandler(service)

	http.HandleFunc("/projects", enableCORS(handler.GetAllProjects))
	http.HandleFunc("/project", enableCORS(handler.GetProjectById))
	http.HandleFunc("/project/create", enableCORS(handler.CreateProject))
	http.HandleFunc("/project/update", enableCORS(handler.UpdateProject))
	http.HandleFunc("/project/delete", enableCORS(handler.DeleteProject))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	logger.Printf("Server is starting on port %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.Fatal("Server failed to start: ", err)
	}
}
