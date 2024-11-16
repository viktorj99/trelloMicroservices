package helpers

import (
	"log"
	"net/http"
	"users-service/handlers"
)

func StartHTTPServer() {
	http.HandleFunc("/users/register", enableCORS(handlers.RegisterUser))
	http.HandleFunc("/users/login", enableCORS(handlers.LoginUser))
	http.HandleFunc("/users", enableCORS(handlers.GetAllUsers))
	http.HandleFunc("/users/members", enableCORS(handlers.GetAllUserMembers))
	http.HandleFunc("/users/", enableCORS(handlers.GetUserByID))
	http.HandleFunc("/users/verification", enableCORS(handlers.VerifyCode))
	log.Println("HTTP server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
