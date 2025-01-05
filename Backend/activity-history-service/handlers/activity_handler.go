package handlers

import (
	"activity-history-service/model"
	"activity-history-service/repositories"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func LogActivity(repo *repositories.EventStoreRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var activity model.Activity
		if err := json.NewDecoder(r.Body).Decode(&activity); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		if activity.ProjectID == "" || activity.ManagerID == "" || activity.ActivityType == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		switch activity.ActivityType {
		case "RemoveUser":
			if activity.UserID != "" {
				activity.Description = fmt.Sprintf(
					"Manager %s removed user with ID %s from project %s.",
					activity.ManagerID, activity.UserID, activity.ProjectID,
				)
			}
		case "AddUser":
			if activity.UserID != "" {
				activity.Description = fmt.Sprintf(
					"Manager %s added user with ID %s to project %s.",
					activity.ManagerID, activity.UserID, activity.ProjectID,
				)
			}
		case "CreateTask":
			activity.Description = fmt.Sprintf(
				"Manager %s created a new task in project %s.",
				activity.ManagerID, activity.ProjectID,
			)
		case "UpdateTaskStatus":
			activity.Description = fmt.Sprintf(
				"Manager %s updated the status of a task in project %s.",
				activity.ManagerID, activity.ProjectID,
			)
		case "AddDocument":
			activity.Description = fmt.Sprintf(
				"Manager %s added a document to a task in project %s.",
				activity.ManagerID, activity.ProjectID,
			)
		default:
			activity.Description = fmt.Sprintf(
				"Manager %s performed action '%s' in project %s.",
				activity.ManagerID, activity.ActivityType, activity.ProjectID,
			)
		}

		activity.Timestamp = time.Now()

		if err := repo.LogActivity(fmt.Sprintf("project-%s", activity.ProjectID), activity.ActivityType, activity); err != nil {
			http.Error(w, "Failed to log activity", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "Activity logged successfully")
	}
}

func GetActivitiesByProject(repo *repositories.EventStoreRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		projectID := vars["projectId"]

		activities, err := repo.GetActivitiesByProject(projectID)
		if err != nil {
			http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(activities)
	}
}
