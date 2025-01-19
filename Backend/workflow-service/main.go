package main

import (
	"log"
	"net/http"
	"os"
	"time"
	"workflow/handlers"
	"workflow/repositories"
	"workflow/services"

	"github.com/gorilla/mux"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	// Load environment variables
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	log.Println("Warning: .env file not found or failed to load")
	// }

	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	// Neo4j credentials from environment
	neo4jURI := os.Getenv("NEO4J_URI")
	neo4jUser := os.Getenv("NEO4J_USER")
	neo4jPassword := os.Getenv("NEO4J_PASSWORD")

	if neo4jURI == "" || neo4jUser == "" || neo4jPassword == "" {
		logger.Fatal("Neo4j credentials (NEO4J_URI, NEO4J_USER, NEO4J_PASSWORD) must be set")
	}

	// Connect to Neo4j
	driver, err := neo4j.NewDriver(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		logger.Fatalf("Failed to connect to Neo4j: %v", err)
	}
	defer driver.Close()

	for i := 0; i < 10; i++ {
		if err := driver.VerifyConnectivity(); err == nil {
		  break
		}
		log.Println("Neo4j not ready yet, waiting 2s...")
		time.Sleep(5 * time.Second)
	  }

	// Initialize repository, service, and handler
	workflowRepo := repositories.NewWorkflowRepository(driver)
	workflowService := services.NewWorkflowService(workflowRepo)
	workflowHandler := handlers.NewWorkflowHandler(workflowService)

	// Setup router
	router := mux.NewRouter()
	router.HandleFunc("/workflow/tasks", workflowHandler.CreateTask).Methods("POST")
	router.HandleFunc("/workflow/tasks/{id}", workflowHandler.GetTask).Methods("GET")
	router.HandleFunc("/workflow/dependencies", workflowHandler.CreateDependency).Methods("POST")
	router.HandleFunc("/workflow/tasks", workflowHandler.GetTasks).Methods("GET")
	router.HandleFunc("/workflow/tasks/dependencies/{id}", workflowHandler.GetTasksWithDependencies).Methods("GET")
	router.HandleFunc("/workflow/dependencies", workflowHandler.GetAllDependencies).Methods("GET")
	router.HandleFunc("/workflow/tasks/update-status/{id}", workflowHandler.UpdateTaskStatus).Methods("PUT")

	// HTTPS server configuration
	httpsPort := os.Getenv("HTTPS_PORT")
	if httpsPort == "" {
		httpsPort = "8443"
	}
	logger.Printf("HTTPS server is starting on port %s...", httpsPort)

	certFile := "certificates/cert.crt"
	keyFile := "certificates/cert.key"

	if err := http.ListenAndServeTLS(":"+httpsPort, certFile, keyFile, router); err != nil {
		logger.Fatalf("HTTPS server failed to start: %v", err)
	}
}
