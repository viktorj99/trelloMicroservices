package main

import (
	"log"
	"net/http"
	"notifications-service/handlers"
	"notifications-service/helpers"
	"notifications-service/repositories"
	service "notifications-service/services"
	"os"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
)

func main() {
	cleanup := helpers.InitTracer()
	defer cleanup()

	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	db := os.Getenv("CASS_DB")
	if db == "" {
		logger.Fatal("CASS_DB environment variable not set")
	}

	cluster := gocql.NewCluster(db)
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Fatal("Failed to connect to Cassandra:", err)
		return
	}
	defer session.Close()

	createKeyspaceQuery := `CREATE KEYSPACE IF NOT EXISTS trello WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};`
	if err := session.Query(createKeyspaceQuery).Exec(); err != nil {
		logger.Fatal("Error creating keyspace:", err)
	}

	createTableQuery := `CREATE TABLE IF NOT EXISTS trello.notifications_by_month (
		user_id TEXT,
		notification_id UUID,
		year_month TEXT,
		created_at TIMESTAMP,
		message TEXT,
		is_read BOOLEAN,
		PRIMARY KEY (user_id, year_month, created_at, notification_id)
	) WITH CLUSTERING ORDER BY (year_month ASC, created_at DESC ,notification_id ASC);`

	if err := session.Query(createTableQuery).Exec(); err != nil {
		logger.Fatal("Error creating table:", err)
	}

	notificationRepo := repositories.NewNotificationRepository(session)
	notificationService := service.NewNotificationService(notificationRepo)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	router := mux.NewRouter()
	tracer := otel.Tracer("notifications-service/main")

	router.HandleFunc("/notifications/project/add", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "POST /notifications/project/add")
		defer span.End()
		notificationHandler.NotifyMembersHandler(w, r.WithContext(ctx))
	}).Methods("POST")

	router.HandleFunc("/notifications/project/remove", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "POST /notifications/project/remove")
		defer span.End()
		notificationHandler.NotifyMembersHandler(w, r.WithContext(ctx))
	}).Methods("POST")

	router.HandleFunc("/notifications/task/add", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "POST /notifications/task/add")
		defer span.End()
		notificationHandler.NotifyMembersHandler(w, r.WithContext(ctx))
	}).Methods("POST")

	router.HandleFunc("/notifications/task/remove", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "POST /notifications/task/remove")
		defer span.End()
		notificationHandler.NotifyMembersHandler(w, r.WithContext(ctx))
	}).Methods("POST")

	router.HandleFunc("/notifications/task/status", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "POST /notifications/task/status")
		defer span.End()
		notificationHandler.NotifyMembersHandler(w, r.WithContext(ctx))
	}).Methods("POST")

	router.HandleFunc("/notifications/by_month", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "GET /notifications/by_month")
		defer span.End()
		notificationHandler.GetNotificationsByMonthHandler(w, r.WithContext(ctx))
	}).Methods("GET")

	router.HandleFunc("/notifications/user", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "GET /notifications/user")
		defer span.End()
		notificationHandler.GetAllNotificationsHandler(w, r.WithContext(ctx))
	}).Methods("GET")

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8084"
	}

	httpsPort := os.Getenv("HTTPS_PORT")
	if httpsPort == "" {
		httpsPort = "8443"
	}

	go func() {
		logger.Printf("HTTP server is starting on port %s...", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			logger.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	logger.Printf("HTTPS server is starting on port %s...", httpsPort)
	if err := http.ListenAndServeTLS(":"+httpsPort, "certificates/cert.crt", "certificates/cert.key", router); err != nil {
		logger.Fatalf("HTTPS server failed to start: %v", err)
	}
}
