package services

import (
	"time"

	"notifications-service/model"
	"notifications-service/repositories"

	"github.com/gocql/gocql"
)

type NotificationService struct {
	repo *repositories.NotificationRepository
}

// NewNotificationService initializes the service layer
func NewNotificationService(repo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// CreateNotification creates a new notification
func (service *NotificationService) CreateNotification(userID gocql.UUID, message string) error {
	now := time.Now()

	notification := model.Notification{
		UserID:         userID,
		NotificationID: gocql.TimeUUID(),      // Generate unique UUID
		YearMonth:      now.Format("2006-01"), // Format time as "YYYY-MM"
		CreatedAt:      now,
		Message:        message,
		IsRead:         false,
	}

	return service.repo.SaveNotification(notification)
}

// GetNotifications gets all notifications for a user in a specific month
func (service *NotificationService) GetNotifications(userID gocql.UUID, yearMonth string) ([]model.Notification, error) {
	return service.repo.GetNotificationsByMonth(userID, yearMonth)
}
