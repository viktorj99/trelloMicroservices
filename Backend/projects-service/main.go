package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"projects-service/auth"

	"projects-service/client"
	"projects-service/handlers"
	"projects-service/helpers"
	"projects-service/repositories"
	"projects-service/services"

	"github.com/gorilla/mux"
)


func main() {
	helpers.LoadingEnv()

	ctx := context.Background()
	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	// Povezivanje sa MongoDB
	dburi := os.Getenv("MONGO_DB_URI")
	mongoClient, err := helpers.SetupMongoClient(ctx, dburi)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB: ", err)
	}
	defer mongoClient.Disconnect(ctx)

	// Povezivanje sa gRPC user-service preko novog klijenta
	userServiceAddress := "localhost:50051"
	userClient, err := client.NewUserClient(userServiceAddress)
	if err != nil {
		logger.Fatal("Failed to connect to user-service: ", err)
	}
	defer userClient.Close()

	// Kreiranje repozitorijuma i servisa
	repo, err := repositories.NewProjectRepo(ctx, logger, "project")
	if err != nil {
		logger.Fatal("Failed to create repository: ", err)
	}
	service := services.NewProjectService(repo)
	handler := handlers.NewProjectHandler(service)
	// Postavljanje ruta za Project REST API
	r := mux.NewRouter()

	r.Handle("/projects", auth.EnableBoth(http.HandlerFunc(handler.GetAllProjects))).Methods("GET")
	r.Handle("/project/{id}", auth.EnableBoth(http.HandlerFunc(handler.GetProjectById))).Methods("GET")
	r.Handle("/project/create", auth.EnableManager(http.HandlerFunc(handler.CreateProject))).Methods("POST")
	r.Handle("/project/{id}", auth.EnableManager(http.HandlerFunc(handler.UpdateProject))).Methods("PUT")
	r.HandleFunc("/project/{id}", handler.UpdateProject).Methods("PUT")
	r.HandleFunc("/project/{id}", handler.DeleteProject).Methods("DELETE")

	// Dodavanje rute za gRPC poziv user-service
	r.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetUsersHandler(w, r, userClient.Client)
	}).Methods("GET")

	// Dodavanje CORS middleware-a
	http.Handle("/", helpers.EnableCORS(r))

	// Pokretanje servera
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
