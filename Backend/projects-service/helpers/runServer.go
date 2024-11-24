package helpers

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func RunServer(router *mux.Router, logger *log.Logger) {
	http.Handle("/", router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8443"
	}
	logger.Printf("HTTPS server is starting on port %s...", port)
	err := http.ListenAndServeTLS(":"+port, "certificates/cert.crt", "certificates/cert.key", nil)
	if err != nil {
		logger.Fatal("Server failed to start: ", err)
	}
}
