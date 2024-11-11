package helpers

import (
	"context"
	"log"
	"projects-service/repositories"
	"projects-service/services"
)

func InitializeService(ctx context.Context, logger *log.Logger) ( *services.ProjectService) {
	repo, err := repositories.NewProjectRepo(ctx, logger, "project")
	if err != nil {
		logger.Fatal("Failed to create repository: ", err)
	}
	service := services.NewProjectService(repo)
	return service
}