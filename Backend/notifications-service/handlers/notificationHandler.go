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

	log.Printf("Received request: ProjectName: %s, UserIDs: %v", req.ProjectName, req.UserIDs)

	// Check if project name and user IDs are provided
	if req.ProjectName == "" || len(req.UserIDs) == 0 {
		http.Error(w, "Project name and user IDs are required", http.StatusBadRequest)
		return
	}

	// Process each user ID and create a notification
	for _, userID := range req.UserIDs {
		// Generate a new UUID for the notification ID
		notificationID, err := gocql.RandomUUID()
		if err != nil {
			log.Printf("Failed to generate notification ID: %v", err)
			http.Error(w, "Failed to create notification ID", http.StatusInternalServerError)
			return
		}

		var message string
		switch r.URL.Path {
		case "/notifications/project/add":
			message = "You have been added to the project: " + req.ProjectName
		case "/notifications/project/remove":
			message = "You have been removed from the project: " + req.ProjectName
		case "/notifications/task/add":
			message = "You have been removed from a Task in the project: " + req.ProjectName
		case "/notifications/task/remove":
			message = "You have been removed from a Task in the project: " + req.ProjectName
		case "/notifications/task/status":
			message = "The status of a task you are assigned to has changed in the project: " + req.ProjectName
		default:
			http.Error(w, "Unknown action", http.StatusBadRequest)
			return
		}

		// Save the notification using the service
		if err := h.notificationService.CreateNotification(userID, notificationID, message); err != nil {
			log.Printf("Failed to create notification for user: %s, error: %v", userID, err)
			http.Error(w, "Failed to create notifications", http.StatusInternalServerError)
			return
		}
	}

	// Return a success response
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Notifications created successfully"))
}

// GetNotificationsByMonthHandler handles the retrieval of notifications for a user in a specific month
func (h *NotificationHandler) GetNotificationsByMonthHandler(w http.ResponseWriter, r *http.Request) {

	// Extract userID and yearMonth from the query parameters
	userID := r.URL.Query().Get("user_id")
	yearMonth := r.URL.Query().Get("year_month")

	if userID == "" || yearMonth == "" {
		http.Error(w, "User ID and Year-Month are required", http.StatusBadRequest)
		return
	}

	// Fetch notifications using the service
	notifications, err := h.notificationService.GetNotifications(userID, yearMonth)
	if err != nil {
		log.Printf("Failed to fetch notifications for user: %s, error: %v", userID, err)
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	// Return the notifications in JSON format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// GetAllNotificationsHandler handles the retrieval of all notifications for a user
func (h *NotificationHandler) GetAllNotificationsHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Request URL: %s", r.URL.String())

	// Extract userID from the query parameters
	userID := r.Header.Get("user_id")

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Fetch all notifications for the user
	notifications, err := h.notificationService.GetAllNotifications(userID)
	if err != nil {
		log.Printf("Failed to fetch notifications for user: %s, error: %v", userID, err)
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	// Return the notifications in JSON format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}
