package helpers

import (
	"log"
	"net/http"
	"strings"
	"users-service/handlers"
)

func StartHTTPServer() {
	http.HandleFunc("/users/register", enableCORS(handlers.RegisterUser))
	http.HandleFunc("/users/login", enableCORS(handlers.LoginUser))
	http.HandleFunc("/users/members", enableCORS(handlers.GetAllUserMembers))
	http.HandleFunc("/users/verification", enableCORS(handlers.VerifyCode))

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/users/") && len(r.URL.Path) > len("/users/") {
			enableCORS(handlers.GetUserByID)(w, r)
		} else {
			enableCORS(handlers.GetAllUsers)(w, r)
		}
	})

	log.Println("HTTP server is running on port 8079...")
	log.Fatal(http.ListenAndServe(":8079", nil))
}
