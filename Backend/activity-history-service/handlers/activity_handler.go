package handlers

import (
	"activity-history-service/model"
	"activity-history-service/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type ActivityHandler struct {
	Service *services.ActivityService
}

func NewActivityHandler(s *services.ActivityService) *ActivityHandler {
	return &ActivityHandler{Service: s}
}

func (h *ActivityHandler) LogActivity(w http.ResponseWriter, r *http.Request) {
	var activity model.Activity
	if err := json.NewDecoder(r.Body).Decode(&activity); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if activity.ProjectID == "" || activity.ActivityType == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	log.Printf("Received activity: %+v\n", activity)

	activity.Timestamp = time.Now().UTC()

	if err := h.Service.LogActivity(activity); err != nil {
		http.Error(w, "Failed to log activity", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Activity logged successfully")
}

func (h *ActivityHandler) GetActivitiesByProject(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["projectId"]
	activities, err := h.Service.GetActivitiesByProject(projectID)
	if err != nil {
		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(activities)
}

func (h *ActivityHandler) GetActivitiesByUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userId"]
	activities, err := h.Service.GetActivitiesByUser(userID)
	if err != nil {
		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(activities)
}
