package main

import (
	"log"
	"users-service/database"
	"users-service/helpers"
	"users-service/repositories"
	"users-service/services"
)

func main() {
	// err := utils.LoadCommonPasswordsToConsul("/app/config/common_passwords.txt")
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
	go helpers.StartHTTPServer() // Pokrenuti i HTTP i HTTPS
	helpers.StartGRPCServer()    // gRPC server na portu 50051
}
