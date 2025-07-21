package helpers

import (
	"log"
	"net/http"
	"strings"
	"users-service/handlers"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func StartHTTPServer() {
	http.HandleFunc("/users/register", otelhttp.NewHandler(http.HandlerFunc(handlers.RegisterUser), "RegisterUser").ServeHTTP)
	http.HandleFunc("/users/login", otelhttp.NewHandler(http.HandlerFunc(handlers.LoginUser), "LoginUser").ServeHTTP)
	http.HandleFunc("/users/members", otelhttp.NewHandler(http.HandlerFunc(handlers.GetAllUserMembers), "GetAllUserMembers").ServeHTTP)
	http.HandleFunc("/users/verification", otelhttp.NewHandler(http.HandlerFunc(handlers.VerifyCode), "VerifyCode").ServeHTTP)
	http.HandleFunc("/users/", otelhttp.NewHandler(http.HandlerFunc(handlers.GetUserByID), "GetUserByID").ServeHTTP)
	http.HandleFunc("/users/delete/", otelhttp.NewHandler(http.HandlerFunc(handlers.DeleteUserById), "DeleteUserById").ServeHTTP)
	http.HandleFunc("/users/forgot-password", otelhttp.NewHandler(http.HandlerFunc(handlers.ForgotPassword), "ForgotPassword").ServeHTTP)
	http.HandleFunc("/users/forgot-password/change", otelhttp.NewHandler(http.HandlerFunc(handlers.ChangeForgotPassword), "ChangeForgotPassword").ServeHTTP)
	http.HandleFunc("/users/magic-link/request", otelhttp.NewHandler(http.HandlerFunc(handlers.RequestMagicLinkHandler), "RequestMagicLinkHandler").ServeHTTP)
	http.HandleFunc("/users/magic-link/login", otelhttp.NewHandler(http.HandlerFunc(handlers.MagicLoginHandler), "MagicLoginHandler").ServeHTTP)
	http.HandleFunc("/users/change-password", otelhttp.NewHandler(http.HandlerFunc(handlers.ChangePassword), "ChangePassword").ServeHTTP)

	http.HandleFunc("/users", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/users/") && len(r.URL.Path) > len("/users/") {
			handlers.GetUserByID(w, r)
		} else {
			handlers.GetAllUsers(w, r)
		}
	}), "DefaultUserHandler").ServeHTTP)

	go func() {
		log.Println("HTTP server is running on port 8080...")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	log.Println("HTTPS server is running on port 8443...")
	log.Fatal(http.ListenAndServeTLS(":8443", "certificates/cert.crt", "certificates/cert.key", nil))
}
