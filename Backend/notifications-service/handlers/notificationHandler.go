package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	service "notifications-service/services"

	"github.com/gocql/gocql"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

// NewNotificationHandler initializes a new NotificationHandler
func NewNotificationHandler(service *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: service}
}

// NotificationRequest represents the request body structure
type NotificationRequest struct {
	ProjectName string   `json:"project_name"`
	UserIDs     []string `json:"user_ids"`
}

// NotifyMembersHandler handles the creation of notifications for multiple members
func (h *NotificationHandler) NotifyMembersHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if project name and user IDs are provided
	if req.ProjectName == "" || len(req.UserIDs) == 0 {
		http.Error(w, "Project name and user IDs are required", http.StatusBadRequest)
		return
	}

	// Process each user ID and create a notification
	for _, id := range req.UserIDs {
		// Convert string ID to gocql.UUID
		userID, err := gocql.ParseUUID(id)
		if err != nil {
			log.Printf("Invalid UUID for user: %s", id)
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}

		// Construct the notification message
		message := "You have been added to the project: " + req.ProjectName

		// Save the notification using the service
		if err := h.notificationService.CreateNotification(userID, message); err != nil {
			log.Printf("Failed to create notification for user: %s, error: %v", userID, err)
			http.Error(w, "Failed to create notifications", http.StatusInternalServerError)
			return
		}
	}

	// Return a success response
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Notifications created successfully"))
}
