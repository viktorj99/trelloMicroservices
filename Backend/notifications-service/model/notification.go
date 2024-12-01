package model

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gocql/gocql"
)

type Notification struct {
	UserID         string     `json:"user_id" validate:"required,uuid4"`
	NotificationID gocql.UUID `json:"notification_id" validate:"required"`
	YearMonth      string     `json:"year_month" validate:"required,datetime=2006-01"`
	CreatedAt      time.Time  `json:"created_at" validate:"required"`
	Message        string     `json:"message" validate:"required,min=5,max=500"`
	IsRead         bool       `json:"is_read"`
}

// Validator function for Notification
func ValidateNotification(notification *Notification) error {
	validate := validator.New()
	return validate.Struct(notification)
}
