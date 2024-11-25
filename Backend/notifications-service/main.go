package main

import (
	"log"
	"net/http"
	"notifications-service/handlers"
	"notifications-service/repositories"
	service "notifications-service/services"
	"os"
	"time"

	"github.com/gocql/gocql"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	db := os.Getenv("CASS_DB")
	// Create Cassandra session
	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "trello"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal("Failed to connect to Cassandra:", err)
		return
	}
	defer func() {
		if session != nil {
			session.Close()
		}
	}()

	// Create keyspace if it doesn't exist
	createKeyspaceQuery := `CREATE KEYSPACE IF NOT EXISTS trello
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};`
	if err := session.Query(createKeyspaceQuery).Exec(); err != nil {
		log.Fatal("Error creating keyspace:", err)
	}

	// Create table if it doesn't exist
	createTableQuery := `CREATE TABLE IF NOT EXISTS trello.notifications_by_month (
		user_id UUID,
		notification_id UUID,
		year_month TEXT,
		created_at TIMESTAMP,
		message TEXT,
		is_read BOOLEAN,
		PRIMARY KEY (user_id, year_month, notification_id)
	) WITH CLUSTERING ORDER BY (year_month ASC, notification_id ASC);`

	if err := session.Query(createTableQuery).Exec(); err != nil {
		log.Fatal("Error creating table:", err)
	}

	// Initialize repositories and services
	notificationRepo := repositories.NewNotificationRepository(session)
	notificationService := service.NewNotificationService(notificationRepo)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Set up router
	router := mux.NewRouter()
	router.HandleFunc("/notifications/project/add", notificationHandler.NotifyMembersHandler).Methods("POST")

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	// Start server
	server := http.Server{
		Addr:         ":8080",
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()
}
