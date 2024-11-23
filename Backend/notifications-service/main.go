package main

import (
	"net/http"
	"notifications-service/handlers"
	"notifications-service/repositories"
	service "notifications-service/services"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
)

func main() {
	// Create Cassandra session
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "trello"
	cluster.Consistency = gocql.Quorum
	session, _ := cluster.CreateSession()
	defer session.Close()

	// Initialize repositories and services
	notificationRepo := repositories.NewNotificationRepository(session)
	notificationService := service.NewNotificationService(notificationRepo)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Set up router
	router := mux.NewRouter()
	router.HandleFunc("/notify-members", notificationHandler.NotifyMembersHandler).Methods("POST")

	// Start server
	http.ListenAndServe(":8080", router)
}
