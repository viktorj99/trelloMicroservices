package main

import (
	"log"
	"net/http"
	"users-service/handlers"
)

func main() {
	http.HandleFunc("/register", handlers.RegisterUser)
	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
