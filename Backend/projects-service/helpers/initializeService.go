package helpers

import (
	"context"
	"log"
	"projects-service/client"
	"projects-service/repositories"
	"projects-service/services"
)

func InitializeService(ctx context.Context, logger *log.Logger, natsURL string) *services.ProjectService {
	repo, err := repositories.NewProjectRepo(ctx, logger, "project")
	if err != nil {
		logger.Fatal("Failed to create repository: ", err)
	}

	natsClient := client.NewNATSClient(natsURL)

	service := services.NewProjectService(repo, natsClient.Conn)
	return service
}
