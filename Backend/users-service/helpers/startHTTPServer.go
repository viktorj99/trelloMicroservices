package helpers

import (
	"log"
	"net/http"
	"users-service/handlers"
)

func StartHTTPServer() {
	http.HandleFunc("/register", enableCORS(handlers.RegisterUser))
	http.HandleFunc("/users", enableCORS(handlers.GetAllUsers))
	http.HandleFunc("/users/", enableCORS(handlers.GetUserByID))
	log.Println("HTTP server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}