package main

import (
	"log"
	"users-service/database"
	"users-service/helpers"
	"users-service/repositories"
	"users-service/services"
)

func main() {
	// err := utils.LoadCommonPasswordsToConsul("config/common_passwords.txt")
	// if err != nil {
	// 	log.Fatalf("Failed to load common passwords to Consul: %v", err)
	// }

	client, err := database.GetMongoClient()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	repositories.InitRepository(client)

	go services.StartKeyExpirationListener()

	// Pokretanje HTTP i gRPC servera paralelno
	go helpers.StartHTTPServer() // HTTP server na portu 8080
	helpers.StartGRPCServer()    // gRPC server na portu 50051
}
