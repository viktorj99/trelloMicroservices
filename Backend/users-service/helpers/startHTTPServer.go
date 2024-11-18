package helpers

import (
	"log"
	"net/http"
	"users-service/handlers"
)

func StartHTTPServer() {
	http.HandleFunc("/register", enableCORS(handlers.RegisterUser))
	http.HandleFunc("/login", enableCORS(handlers.LoginUser))
	http.HandleFunc("/users", enableCORS(handlers.GetAllUsers))
	http.HandleFunc("/users/members", enableCORS(handlers.GetAllUserMembers))
	http.HandleFunc("/users/", enableCORS(handlers.GetUserByID))
	http.HandleFunc("/verification", enableCORS(handlers.VerifyCode))
	http.HandleFunc("/forgot-password", enableCORS(handlers.ForgotPassword))
	http.HandleFunc("/change-password", enableCORS(handlers.ChangePassword))
	http.HandleFunc("/magic-link/request", enableCORS(handlers.RequestMagicLinkHandler)) // Add this
	http.HandleFunc("/magic-link/login", enableCORS(handlers.MagicLoginHandler))
	log.Println("HTTP server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
