package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"projects-service/model"
	"projects-service/services"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProjectHandler struct {
	service  *services.ProjectService
	natsConn *nats.Conn
}

func NewProjectHandler(service *services.ProjectService, natsConn *nats.Conn) *ProjectHandler {
	return &ProjectHandler{
		service:  service,
		natsConn: natsConn,
	}
}

func (ph *ProjectHandler) GetAllProjects(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	projects, err := ph.service.GetAllProjects(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (ph *ProjectHandler) GetProjectById(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	project, err := ph.service.GetProjectById(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (ph *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	var project model.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if project.MinMembers <= 0 {
		http.Error(w, "Minimum members must be greater than 0", http.StatusBadRequest)
		return
	}

	if project.MaxMembers < project.MinMembers {
		http.Error(w, "Maximum members cannot be less than minimum members", http.StatusBadRequest)
		return
	}

	if len(project.Members) < project.MinMembers || len(project.Members) > project.MaxMembers {
		http.Error(w, "The number of members must be between the minimum and maximum limits", http.StatusBadRequest)
		return
	}

	result, err := ph.service.CreateProject(ctx, &project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (ph *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var updateData bson.M
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	result, err := ph.service.UpdateProject(ctx, id, updateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// NATS publishing logic (assuming new member was added in the "members" field)
	if members, ok := updateData["members"].([]interface{}); ok {
		// Assuming the new member is the last one in the list
		if len(members) > 0 {
			newMember := members[len(members)-1].(map[string]interface{})
			userID := newMember["id"].(string) // Adjust depending on your data structure

			// Publish the userID to NATS
			err = ph.publishUserIDToNATS(userID)
			if err != nil {
				log.Println("Failed to publish user ID to NATS:", err)
			} else {
				log.Println("Successfully published user ID to NATS:", userID)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// NATS Publishing
func (ph *ProjectHandler) publishUserIDToNATS(userID string) error {

	natsConn := ph.natsConn
	if natsConn == nil {
		return errors.New("NATS connection is not initialized")
	}

	// Publishing the userID
	err := natsConn.Publish("user.added", []byte(userID))
	return err
}

func (ph *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	result, err := ph.service.DeleteProject(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
