package helpers

import (
	"log"
	"net/http"
	"strings"
	"users-service/handlers"
)

func StartHTTPServer() {
	go func() {
		log.Println("HTTP server is running on port 8080...")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	http.HandleFunc("/users/register", handlers.RegisterUser)
	http.HandleFunc("/users/login", handlers.LoginUser)
	http.HandleFunc("/users/members", handlers.GetAllUserMembers)
	http.HandleFunc("/users/verification", handlers.VerifyCode)
	http.HandleFunc("/users/", handlers.GetUserByID)
	http.HandleFunc("/users/delete/", handlers.DeleteUserById)
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

	log.Println("HTTPS server is running on port 8443...")
	log.Fatal(http.ListenAndServeTLS(":8443", "certificates/cert.crt", "certificates/cert.key", nil))
}
