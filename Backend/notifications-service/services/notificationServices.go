package services

import (
	"context"
	"time"

	"notifications-service/model"
	"notifications-service/repositories"

	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type NotificationService struct {
	repo *repositories.NotificationRepository
}

func NewNotificationService(repo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (service *NotificationService) CreateNotification(userID string, notificationID gocql.UUID, message string) error {
	_, span := otel.Tracer("notifications-service/services").Start(context.Background(), "CreateNotification")
	defer span.End()

	span.SetAttributes(
		attribute.String("userID", userID),
		attribute.String("notificationID", notificationID.String()),
		attribute.String("message", message),
	)

	now := time.Now()
	notification := model.Notification{
		UserID:         userID,
		NotificationID: notificationID,
		YearMonth:      now.Format("2006-01"),
		CreatedAt:      now,
		Message:        message,
		IsRead:         false,
	}

	err := service.repo.SaveNotification(notification)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (service *NotificationService) GetNotifications(userID string, yearMonth string) ([]model.Notification, error) {
	_, span := otel.Tracer("notifications-service/services").Start(context.Background(), "GetNotificationsService")
	defer span.End()

	span.SetAttributes(
		attribute.String("userID", userID),
		attribute.String("yearMonth", yearMonth),
	)

	notifications, err := service.repo.GetNotificationsByMonth(userID, yearMonth)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return notifications, nil
}

func (service *NotificationService) GetAllNotifications(userID string) ([]model.Notification, error) {
	_, span := otel.Tracer("notifications-service/services").Start(context.Background(), "GetAllNotificationsService")
	defer span.End()

	span.SetAttributes(attribute.String("userID", userID))

	notifications, err := service.repo.GetAllNotifications(userID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return notifications, nil
}
