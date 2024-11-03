package main

import (
	"log"
	"net/http"
	"users-service/database"
	"users-service/handlers"
	"users-service/repositories"
)

func enableCORS(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handlerFunc(w, r)
	}
}

func main() {
	client, err := database.GetMongoClient()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	repositories.InitRepository(client)

	http.HandleFunc("/register", enableCORS(handlers.RegisterUser))

	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}