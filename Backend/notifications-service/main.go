package main

import (
	"log"
	"net/http"

	"notifications-service/handlers"
	"notifications-service/repositories"
	service "notifications-service/services"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
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

	natsConn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer natsConn.Close()

	// Initialize the NotificationHandler with NATS connection
	notificationHandler := handlers.NewNotificationHandler(notificationService, natsConn)

	r := mux.NewRouter()

	r.HandleFunc("/notifications", notificationHandler.CreateNotificationHandler).Methods("POST")
	r.HandleFunc("/notifications", notificationHandler.GetNotificationsHandler).Methods("GET")

	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
