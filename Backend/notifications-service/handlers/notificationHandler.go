package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	service "notifications-service/services"

	"github.com/gocql/gocql"
	"github.com/nats-io/nats.go"
)

type NotificationHandler struct {
	service  *service.NotificationService
	natsConn *nats.Conn
}

// NewNotificationHandler initializes the handler with the service layer and NATS connection
func NewNotificationHandler(service *service.NotificationService, natsConn *nats.Conn) *NotificationHandler {
	handler := &NotificationHandler{
		service:  service,
		natsConn: natsConn,
	}

	handler.subscribeToNotifications()

	return handler
}

// subscribeToNotifications subscribes to the "notifications" subject on NATS
func (h *NotificationHandler) subscribeToNotifications() {
	subject := "notifications" // Define the subject to subscribe to
	_, err := h.natsConn.Subscribe(subject, func(msg *nats.Msg) {
		var notification struct {
			UserID  string `json:"user_id"`
			Message string `json:"message"`
		}

		// Decode the incoming message
		if err := json.Unmarshal(msg.Data, &notification); err != nil {
			log.Println("Failed to unmarshal notification:", err)
			return
		}

		// Parse the UserID to gocql.UUID
		userID, err := gocql.ParseUUID(notification.UserID)
		if err != nil {
			log.Println("Invalid user_id in NATS message:", err)
			return
		}

		// Create the notification using the service
		if err := h.service.CreateNotification(userID, notification.Message); err != nil {
			log.Println("Failed to create notification:", err)
			return
		}

		log.Printf("Notification created for user %s: %s", notification.UserID, notification.Message)
	})

	if err != nil {
		log.Fatalf("Error subscribing to NATS subject 'notifications': %v", err)
	}
}

// CreateNotificationHandler handles the creation of a notification via HTTP
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
