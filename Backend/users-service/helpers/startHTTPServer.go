package helpers

import (
	"log"
	"net/http"
	"strings"
	"users-service/handlers"
)

func StartHTTPServer() {
	http.HandleFunc("/users/register", handlers.RegisterUser)
	http.HandleFunc("/users/login", handlers.LoginUser)
	http.HandleFunc("/users/members", handlers.GetAllUserMembers)
	http.HandleFunc("/users/verification", handlers.VerifyCode)
	http.HandleFunc("/users/", handlers.GetUserByID)
	http.HandleFunc("/users/forgot-password", handlers.ForgotPassword)
	http.HandleFunc("/users/forgot-password/change", handlers.ChangeForgotPassword)
	http.HandleFunc("/users/magic-link/request", handlers.RequestMagicLinkHandler)
	http.HandleFunc("/users/magic-link/login", handlers.MagicLoginHandler)
	http.HandleFunc("/users/change-password", handlers.ChangePassword)

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/users/") && len(r.URL.Path) > len("/users/") {
			handlers.GetUserByID(w, r)
		} else {
			handlers.GetAllUsers(w, r)
		}
	})

	log.Println("HTTP server is running on port 8079...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
