package handlers

import (
	"encoding/json"
	"net/http"
	service "notifications-service/services"

	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(service *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: service}
}

type NotificationRequest struct {
	ProjectName string   `json:"project_name"`
	UserIDs     []string `json:"user_ids"`
}

func (h *NotificationHandler) NotifyMembersHandler(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("notifications-service/handlers")
	_, span := tracer.Start(r.Context(), "NotifyMembersHandler")
	defer span.End()

	var req NotificationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("project_name", req.ProjectName),
		attribute.StringSlice("user_ids", req.UserIDs),
	)

	if req.ProjectName == "" || len(req.UserIDs) == 0 {
		http.Error(w, "Project name and user IDs are required", http.StatusBadRequest)
		return
	}

	for _, userID := range req.UserIDs {
		notificationID, err := gocql.RandomUUID()
		if err != nil {
			span.RecordError(err)
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
			message = "You have been assigned to a Task in the project: " + req.ProjectName
		case "/notifications/task/remove":
			message = "You have been removed from a Task in the project: " + req.ProjectName
		case "/notifications/task/status":
			message = "The Status of a Task you are assigned on has changed in the project: " + req.ProjectName
		default:
			http.Error(w, "Unknown action", http.StatusBadRequest)
			return
		}

		if err := h.notificationService.CreateNotification(userID, notificationID, message); err != nil {
			span.RecordError(err)
			http.Error(w, "Failed to create notifications", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Notifications created successfully"))
}

func (h *NotificationHandler) GetNotificationsByMonthHandler(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("notifications-service/handlers")
	_, span := tracer.Start(r.Context(), "GetNotificationsByMonthHandler")
	defer span.End()

	userID := r.URL.Query().Get("user_id")
	yearMonth := r.URL.Query().Get("year_month")

	span.SetAttributes(
		attribute.String("user_id", userID),
		attribute.String("year_month", yearMonth),
	)

	if userID == "" || yearMonth == "" {
		http.Error(w, "User ID and Year-Month are required", http.StatusBadRequest)
		return
	}

	notifications, err := h.notificationService.GetNotifications(userID, yearMonth)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

func (h *NotificationHandler) GetAllNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("notifications-service/handlers")
	_, span := tracer.Start(r.Context(), "GetAllNotificationsHandler")
	defer span.End()

	userID := r.Header.Get("user_id")

	span.SetAttributes(attribute.String("user_id", userID))

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	notifications, err := h.notificationService.GetAllNotifications(userID)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}
