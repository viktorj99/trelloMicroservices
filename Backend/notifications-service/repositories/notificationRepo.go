package repositories

import (
	"notifications-service/model"

	"github.com/gocql/gocql"
)

type NotificationRepository struct {
	session *gocql.Session
}

// NewNotificationRepository initializes the repository
func NewNotificationRepository(session *gocql.Session) *NotificationRepository {
	return &NotificationRepository{session: session}
}

// SaveNotification saves a notification to Cassandra
func (repo *NotificationRepository) SaveNotification(notification model.Notification) error {
	query := `INSERT INTO notifications_by_month (user_id, notification_id, year_month, created_at, message, is_read)
		VALUES (?, ?, ?, ?, ?, ?)`

	if err := repo.session.Query(query,
		notification.UserID,
		notification.NotificationID,
		notification.YearMonth,
		notification.CreatedAt,
		notification.Message,
		notification.IsRead).Exec(); err != nil {
		return err
	}

	return nil
}

// GetNotificationsByMonth retrieves all notifications for a user in a given month
func (repo *NotificationRepository) GetNotificationsByMonth(userID gocql.UUID, yearMonth string) ([]model.Notification, error) {
	var notifications []model.Notification
	query := `SELECT user_id, notification_id, year_month, created_at, message, is_read
		FROM notifications_by_month WHERE user_id = ? AND year_month = ?`

	iter := repo.session.Query(query, userID, yearMonth).Iter()
	var notification model.Notification
	for iter.Scan(&notification.UserID, &notification.NotificationID, &notification.YearMonth, &notification.CreatedAt, &notification.Message, &notification.IsRead) {
		notifications = append(notifications, notification)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return notifications, nil
}
