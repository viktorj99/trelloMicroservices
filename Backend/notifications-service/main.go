package main

import (
	"log"
	"net/http"

	"notifications-service/handlers"
	"notifications-service/repositories"
	service "notifications-service/services"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize Cassandra session
	cluster := gocql.NewCluster("127.0.0.1") // Add Cassandra nodes
	cluster.Keyspace = "notifications"
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal("Failed to connect to Cassandra:", err)
	}
	defer session.Close()

	// Initialize repository, service, and handler
	notificationRepo := repositories.NewNotificationRepository(session)
	notificationService := service.NewNotificationService(notificationRepo)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Set up the router
	r := mux.NewRouter()

	// Define routes and assign handlers
	r.HandleFunc("/notifications", notificationHandler.CreateNotificationHandler).Methods("POST")
	r.HandleFunc("/notifications", notificationHandler.GetNotificationsHandler).Methods("GET")

	// Start the server
	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
