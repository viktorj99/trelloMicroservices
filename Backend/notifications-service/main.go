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
	router.HandleFunc("/notifications/by_month", notificationHandler.GetNotificationsByMonthHandler).Methods("GET")
	router.HandleFunc("/notifications/user", notificationHandler.GetAllNotificationsHandler).Methods("GET")

	// Start server
	// server := http.Server{
	// 	Addr:         ":8080",
	// 	IdleTimeout:  120 * time.Second,
	// 	ReadTimeout:  1 * time.Second,
	// 	WriteTimeout: 1 * time.Second,
	// }

	// go func() {
	// 	err := server.ListenAndServe()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	logger.Printf("Server is starting on port %s", port)

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		logger.Fatal("Server failed to start: ", err)
	}

}
