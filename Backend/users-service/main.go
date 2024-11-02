package main

import (
	"log"
	"net/http"
	"users-service/database"
	"users-service/handlers"
	"users-service/repositories"
)

func main() {
	client, err := database.GetMongoClient()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	repositories.InitRepository(client)

	http.HandleFunc("/register", handlers.RegisterUser)

	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
