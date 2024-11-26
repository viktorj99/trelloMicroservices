package model

import (
	"time"

	"github.com/gocql/gocql"
)

type Notification struct {
	UserID         string     `json:"user_id"`
	NotificationID gocql.UUID `json:"notification_id"` // Unique notification ID
	YearMonth      string     `json:"year_month"`      // Partition by year and month (e.g., "2024-11")
	CreatedAt      time.Time  `json:"created_at"`      // Timestamp when notification was created
	Message        string     `json:"message"`
	IsRead         bool       `json:"is_read"`
}
