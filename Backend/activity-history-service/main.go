package main

import (
	"log"
	"net/http"
	"os"

	"activity-history-service/handlers"
	"activity-history-service/repositories"

	"github.com/gorilla/mux"
)

func main() {
	eventStoreURI := os.Getenv("EVENTSTORE_URI")
	if eventStoreURI == "" {
		log.Fatal("EVENTSTORE_URI is not set")
	}

	repo, err := repositories.NewEventStoreRepository(eventStoreURI)
	if err != nil {
		log.Fatalf("Failed to connect to EventStoreDB: %v", err)
	}

	activityHandler := handlers.LogActivity(repo)
	getActivitiesHandler := handlers.GetActivitiesByProject(repo)

	router := mux.NewRouter()
	router.HandleFunc("/activities/create", activityHandler).Methods("POST")
	router.HandleFunc("/activities/project/{projectId}", getActivitiesHandler).Methods("GET")

	errChan := make(chan error)

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	go func() {
		log.Printf("HTTP server is starting on port %s...", httpPort)
		errChan <- http.ListenAndServe(":"+httpPort, router)
	}()

	httpsPort := os.Getenv("HTTPS_PORT")
	if httpsPort == "" {
		httpsPort = "8443"
	}
	go func() {
		log.Printf("HTTPS server is starting on port %s...", httpsPort)
		errChan <- http.ListenAndServeTLS(":"+httpsPort, "certificates/cert.crt", "certificates/cert.key", router)
	}()

	err = <-errChan
	log.Fatalf("Server error: %v", err)
}
