package main

import (
	"log"
	"net/http"
	"notifications-service/handlers"
	"notifications-service/repositories"
	service "notifications-service/services"
	"os"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
)

func main() {

	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	db := os.Getenv("CASS_DB")
	// Create Cassandra session
	cluster := gocql.NewCluster(db)
	//cluster.Keyspace = "trello"
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
	createKeyspaceQuery := `CREATE KEYSPACE IF NOT EXISTS trello WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};`
	if err := session.Query(createKeyspaceQuery).Exec(); err != nil {
		log.Fatal("Error creating keyspace:", err)
	}

	// Create table if it doesn't exist
	createTableQuery := `CREATE TABLE IF NOT EXISTS trello.notifications_by_month (
		user_id TEXT,
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
	router.HandleFunc("/notifications/project/remove", notificationHandler.NotifyMembersHandler).Methods("POST")
	router.HandleFunc("/notifications/task/add", notificationHandler.NotifyMembersHandler).Methods("POST")
	router.HandleFunc("/notifications/task/remove", notificationHandler.NotifyMembersHandler).Methods("POST")
	router.HandleFunc("/notifications/task/status", notificationHandler.NotifyMembersHandler).Methods("POST")

	router.HandleFunc("/notifications/by_month", notificationHandler.GetNotificationsByMonthHandler).Methods("GET")
	router.HandleFunc("/notifications/user", notificationHandler.GetAllNotificationsHandler).Methods("GET")

	// HTTP and HTTPS ports
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8084" // Default HTTP port
	}

	httpsPort := os.Getenv("HTTPS_PORT")
	if httpsPort == "" {
		httpsPort = "8443" // Default HTTPS port
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Printf("HTTP server is starting on port %s...", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			logger.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	// Start HTTPS server
	logger.Printf("HTTPS server is starting on port %s...", httpsPort)
	if err := http.ListenAndServeTLS(":"+httpsPort, "certificates/cert.crt", "certificates/cert.key", router); err != nil {
		logger.Fatalf("HTTPS server failed to start: %v", err)
	}
}
