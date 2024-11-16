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

	// Retrieve gRPC service address and API Gateway port from environment
	grpcUserService := os.Getenv("USER_SERVICE") // gRPC service address
	port := os.Getenv("PORT")                    // API Gateway port
	if port == "" {
		port = "8080" // Default port if not set
	}

	// Set up a reverse proxy multiplexer
	http.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		proxyToUserService(w, r, grpcUserService)
	})

	// Start the API Gateway HTTP server
	log.Printf("API Gateway is running on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start API Gateway server: %v", err)
	}
}

func proxyToUserService(w http.ResponseWriter, r *http.Request, grpcUserService string) {
	// Remove the "/api" prefix from the URL path
	newPath := strings.TrimPrefix(r.URL.Path, "/api")

	// Construct the new request to the user-service
	r.URL.Path = newPath
	r.Host = grpcUserService

	// Proxy the request to the user-service
	proxy := &http.Transport{}
	resp, err := proxy.RoundTrip(r)
	if err != nil {
		log.Printf("Error proxying request: %v", err)
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
	_, err = io.Copy(w, resp.Body) // Correctly copy the response body
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

func loadEnv() error {
	_, err := os.Stat(".env")
	if os.IsNotExist(err) {
		return nil
	}
	return godotenv.Load()
}
