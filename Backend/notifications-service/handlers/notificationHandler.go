package handlers

import (
	"encoding/json"
	"net/http"
	service "notifications-service/services"

	"github.com/gocql/gocql"
)

type NotificationHandler struct {
	service *service.NotificationService
}

// NewNotificationHandler initializes the handler with the service layer
func NewNotificationHandler(service *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// CreateNotificationHandler handles the creation of a notification
func (h *NotificationHandler) CreateNotificationHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID  string `json:"user_id"`
		Message string `json:"message"`
	}

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse UserID to gocql.UUID
	userID, err := gocql.ParseUUID(req.UserID)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	// Call the service to create the notification
	if err := h.service.CreateNotification(userID, req.Message); err != nil {
		http.Error(w, "Failed to create notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "notification created"})
}

// GetNotificationsHandler handles retrieving notifications by user and month
func (h *NotificationHandler) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	yearMonth := r.URL.Query().Get("year_month")

	// Parse UserID
	userID, err := gocql.ParseUUID(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	// Call the service to get notifications
	notifications, err := h.service.GetNotifications(userID, yearMonth)
	if err != nil {
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notifications)
}
