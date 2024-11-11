package main

import (
	"log"
	"users-service/database"
	"users-service/helpers"
	"users-service/repositories"
)

func main() {
	client, err := database.GetMongoClient()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	repositories.InitRepository(client)

	// Pokretanje HTTP i gRPC servera paralelno
	go helpers.StartHTTPServer() // HTTP server na portu 8080
	helpers.StartGRPCServer()    // gRPC server na portu 50051
}
