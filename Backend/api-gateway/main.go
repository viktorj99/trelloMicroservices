package main

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	userService := os.Getenv("USER_SERVICE")
	projectService := os.Getenv("PROJECT_SERVICE")
	taskService := os.Getenv("TASK_SERVICE")
	notificationService := os.Getenv("NOTIFICATION_SERVICE")
	workflowService := os.Getenv("WORKFLOW_SERVICE")

	port := os.Getenv("PORT")

	if userService == "" || projectService == "" || taskService == "" || notificationService == "" {
		log.Fatal("One or more service addresses are not set in the environment variables")
	}

	if port == "" {
		port = "8443"
	}

	http.Handle("/api/users/", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, userService)
	})))

	http.Handle("/api/projects/", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, projectService)
	})))

	http.Handle("/api/tasks/", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, taskService)
	})))

	http.Handle("/api/notifications/", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, notificationService)
	})))

	http.Handle("/api/workflow/", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, workflowService)
	})))

	log.Printf("API Gateway is running on HTTPS port %s...", port)
	if err := http.ListenAndServeTLS(":"+port, "certificates/cert.crt", "certificates/cert.key", nil); err != nil {
		log.Fatalf("Failed to start API Gateway HTTPS server: %v", err)
	}
}

func enableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, user_id")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func proxyToService(w http.ResponseWriter, r *http.Request, serviceURL string) {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Skip certificate validation for internal HTTPS
		},
	}

	// Remove the "/api" prefix from the URL path
	if !strings.HasPrefix(serviceURL, "http") {
		serviceURL = "https://" + serviceURL
	}
	newPath := strings.TrimPrefix(r.URL.Path, "/api")
	newPath = strings.TrimSuffix(newPath, "/") // Remove the trailing slash
	fullURL := serviceURL + newPath

	// Log the constructed URL for debugging
	log.Printf("Forwarding request to: %s", fullURL)

	// Create a new HTTP request for the target service
	req, err := http.NewRequest(r.Method, fullURL, r.Body)
	log.Printf("Forwarding to service: %s", fullURL)
	log.Printf("Request Headers: %v", req.Header)
	log.Printf("Request Method: %s", req.Method)

	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding request to %s: %v", fullURL, err)
		http.Error(w, "Failed to forward request: "+err.Error(), http.StatusBadGateway)
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
