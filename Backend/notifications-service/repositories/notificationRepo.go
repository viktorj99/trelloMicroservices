package repositories

import (
	"context"
	"notifications-service/model"

	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type NotificationRepository struct {
	session *gocql.Session
}

func NewNotificationRepository(session *gocql.Session) *NotificationRepository {
	return &NotificationRepository{session: session}
}

func (repo *NotificationRepository) SaveNotification(notification model.Notification) error {
	tracer := otel.Tracer("notifications-service/repositories")
	_, span := tracer.Start(context.Background(), "SaveNotificationRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("userID", notification.UserID),
		attribute.String("notificationID", notification.NotificationID.String()),
		attribute.String("yearMonth", notification.YearMonth),
		attribute.String("message", notification.Message),
	)

	query := `INSERT INTO trello.notifications_by_month (user_id, notification_id, year_month, created_at, message, is_read)
		VALUES (?, ?, ?, ?, ?, ?)`

	if err := repo.session.Query(query,
		notification.UserID,
		notification.NotificationID,
		notification.YearMonth,
		notification.CreatedAt,
		notification.Message,
		notification.IsRead).Exec(); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (repo *NotificationRepository) GetNotificationsByMonth(userID string, yearMonth string) ([]model.Notification, error) {
	tracer := otel.Tracer("notifications-service/repositories")
	_, span := tracer.Start(context.Background(), "GetNotificationsByMonthRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("userID", userID),
		attribute.String("yearMonth", yearMonth),
	)

	var notifications []model.Notification
	query := `SELECT user_id, notification_id, year_month, created_at, message, is_read
		FROM trello.notifications_by_month WHERE user_id = ? AND year_month = ?`

	iter := repo.session.Query(query, userID, yearMonth).Iter()
	var notification model.Notification
	for iter.Scan(&notification.UserID, &notification.NotificationID, &notification.YearMonth, &notification.CreatedAt, &notification.Message, &notification.IsRead) {
		notifications = append(notifications, notification)
	}

	if err := iter.Close(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return notifications, nil
}

func (repo *NotificationRepository) GetAllNotifications(userID string) ([]model.Notification, error) {
	tracer := otel.Tracer("notifications-service/repositories")
	_, span := tracer.Start(context.Background(), "GetAllNotificationsRepo")
	defer span.End()

	span.SetAttributes(attribute.String("userID", userID))

	var notifications []model.Notification
	query := `SELECT user_id, notification_id, year_month, created_at, message, is_read
		FROM trello.notifications_by_month WHERE user_id = ?`

	iter := repo.session.Query(query, userID).Iter()
	var notification model.Notification
	for iter.Scan(&notification.UserID, &notification.NotificationID, &notification.YearMonth, &notification.CreatedAt, &notification.Message, &notification.IsRead) {
		notifications = append(notifications, notification)
	}

	if err := iter.Close(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return notifications, nil
}
