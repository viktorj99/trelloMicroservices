package helpers

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func RunServer(router *mux.Router, logger *log.Logger) {
	http.Handle("/", router)

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	httpsPort := os.Getenv("HTTPS_PORT")
	if httpsPort == "" {
		httpsPort = "8443"
	}

	go func() {
		logger.Printf("HTTP server is starting on port %s...", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			logger.Fatalf("HTTP server failed: %v", err)
		}
	}()

	logger.Printf("HTTPS server is starting on port %s...", httpsPort)
	if err := http.ListenAndServeTLS(":"+httpsPort, "certificates/cert.crt", "certificates/cert.key", nil); err != nil {
		logger.Fatalf("HTTPS server failed: %v", err)
	}
}
