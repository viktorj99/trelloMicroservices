package services

import (
	"activity-history-service/model"
	"activity-history-service/repositories"
	"fmt"
)

type ActivityService struct {
	Repo *repositories.EventStoreRepository
}

func NewActivityService(repo *repositories.EventStoreRepository) *ActivityService {
	return &ActivityService{Repo: repo}
}

func (s *ActivityService) LogActivity(activity model.Activity) error {
	stream := fmt.Sprintf("project-%s", activity.ProjectID)
	return s.Repo.AppendToStream(stream, activity.ActivityType, activity)
}

func (s *ActivityService) GetActivitiesByProject(projectID string) ([]map[string]interface{}, error) {
	return s.Repo.ReadStream(projectID)
}

func (s *ActivityService) GetActivitiesByUser(userID string) ([]map[string]interface{}, error) {
	return s.Repo.ReadStreamByUser(userID)
}
