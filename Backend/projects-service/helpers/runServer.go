package helpers

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func RunServer(router *mux.Router, logger *log.Logger) {
	http.Handle("/", EnableCORS(router))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	logger.Printf("Server is starting on port %s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.Fatal("Server failed to start: ", err)
	}
}