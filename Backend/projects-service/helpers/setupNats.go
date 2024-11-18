package helpers

import (
	"log"
	"projects-service/handlers"
	"projects-service/services"

	"github.com/nats-io/nats.go"
)

// SetupNats initializes the NATS connection and returns a ProjectHandler
func SetupNats(projectService *services.ProjectService) (*handlers.ProjectHandler, *nats.Conn, error) {
	natsConn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
		return nil, nil, err
	}

	// Create a new ProjectHandler with NATS connection
	projectHandler := handlers.NewProjectHandler(projectService, natsConn)

	return projectHandler, natsConn, nil
}
