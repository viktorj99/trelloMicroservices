package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := loadEnv()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Retrieve service addresses and API Gateway port from environment variables
	userService := os.Getenv("USER_SERVICE")
	projectService := os.Getenv("PROJECT_SERVICE")
	taskService := os.Getenv("TASK_SERVICE")
	port := os.Getenv("PORT")

	if userService == "" || projectService == "" || taskService == "" {
		log.Fatal("One or more service addresses are not set in the environment variables")
	}

	if port == "" {
		port = "8080"
	}

	// Set up reverse proxy for each service
	http.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, userService)
	})
	http.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, projectService)
		w.WriteHeader(http.StatusNoContent)
		return
	})
	http.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, taskService)
	})

	// Start the API Gateway HTTP server
	log.Printf("API Gateway is running on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start API Gateway server: %v", err)
	}
}

func handleCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173") // Replace with specific frontend origin or use "*" for all origins
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true") // Needed if credentials (cookies, auth headers) are used
}

func proxyToService(w http.ResponseWriter, r *http.Request, serviceURL string) {
	// Remove the "/api" prefix from the URL path
	newPath := strings.TrimPrefix(r.URL.Path, "/api")
	newPath = strings.TrimSuffix(newPath, "/") // Remove the trailing slash
	fullURL := serviceURL + newPath

	// Log the constructed URL for debugging
	log.Printf("Forwarding request to: %s", fullURL)

	// Create a new HTTP request for the target service
	req, err := http.NewRequest(r.Method, fullURL, r.Body)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	// Perform the request to the target service
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy headers from the response
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write the response status and body
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

func loadEnv() error {
	_, err := os.Stat(".env")
	if os.IsNotExist(err) {
		log.Println(".env file not found")
		return nil
	}
	log.Println("Loading .env file...")
	return godotenv.Load()
}
