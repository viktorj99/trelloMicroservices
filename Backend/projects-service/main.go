package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"projects-service/auth"
	"projects-service/handlers"
	"projects-service/repositories"
	"projects-service/services"

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

	r := mux.NewRouter()

	r.Handle("/projects", auth.EnableBoth(http.HandlerFunc(handler.GetAllProjects))).Methods("GET")
	r.Handle("/project/{id}", auth.EnableBoth(http.HandlerFunc(handler.GetProjectById))).Methods("GET")
	r.Handle("/project/create", auth.EnableManager(http.HandlerFunc(handler.CreateProject))).Methods("POST")
	r.Handle("/project/{id}", auth.EnableManager(http.HandlerFunc(handler.UpdateProject))).Methods("PUT")
	r.Handle("/project/{id}", auth.EnableManager(http.HandlerFunc(handler.DeleteProject))).Methods("DELETE")

	http.Handle("/", enableCORS(r))

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
