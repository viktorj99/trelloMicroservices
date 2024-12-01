package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"projects-service/model"
	"projects-service/services"
	"time"

	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProjectHandler struct {
	service *services.ProjectService
}

func NewProjectHandler(service *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		service: service,
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

	sanitizer := bluemonday.StrictPolicy()
	project.Name = sanitizer.Sanitize(project.Name)

	existingProject, err := ph.service.GetProjectByName(ctx, project.Name)
	if err != nil {
		http.Error(w, "Error checking project name", http.StatusInternalServerError)
		return
	}
	if existingProject != nil {
		http.Error(w, "A project with this name already exists", http.StatusBadRequest)
		return
	}

	if project.ExpectedEndDate.IsZero() || project.ExpectedEndDate.Time.Before(time.Now()) {
		http.Error(w, "Expected end date cannot be empty or in the past", http.StatusBadRequest)
		return
	}

	if project.MinMembers <= 0 {
		http.Error(w, "MinMembers must be greater than 0", http.StatusBadRequest)
		return
	}
	if project.MaxMembers < project.MinMembers {
		http.Error(w, "MaxMembers must be greater than or equal to MinMembers", http.StatusBadRequest)
		return
	}

	if len(project.Members) < project.MinMembers || len(project.Members) > project.MaxMembers {
		http.Error(w, "The number of members must be between MinMembers and MaxMembers", http.StatusBadRequest)
		return
	}

	result, err := ph.service.CreateProject(ctx, &project)
	if err != nil {
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (ph *ProjectHandler) AddMemberToProject(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	vars := mux.Vars(r)
	projectID := vars["id"]

	id, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var member model.User
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		http.Error(w, "Invalid member data", http.StatusBadRequest)
		return
	}

	project, err := ph.service.GetProjectById(ctx, id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	for _, m := range project.Members {
		if m.ID == member.ID {
			http.Error(w, "Member already in project", http.StatusBadRequest)
			return
		}
	}

	if len(project.Members)+1 > project.MaxMembers {
		http.Error(w, "Cannot add member: exceeds maximum limit", http.StatusBadRequest)
		return
	}

	updateData := bson.M{
		"$push": bson.M{"members": member},
	}

	if _, err := ph.service.UpdateProject(ctx, id, updateData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Member added successfully")
}

func (ph *ProjectHandler) RemoveMemberFromProject(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	vars := mux.Vars(r)
	projectID := vars["id"]

	id, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var member model.User
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		http.Error(w, "Invalid member data", http.StatusBadRequest)
		return
	}

	project, err := ph.service.GetProjectById(ctx, id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	memberExists := false
	for _, m := range project.Members {
		if m.ID == member.ID {
			memberExists = true
			break
		}
	}

	if !memberExists {
		http.Error(w, "Member not found in project", http.StatusBadRequest)
		return
	}

	if len(project.Members)-1 < project.MinMembers {
		http.Error(w, "Cannot remove member: falls below minimum limit", http.StatusBadRequest)
		return
	}

	updateData := bson.M{
		"$pull": bson.M{"members": bson.M{"_id": member.ID}},
	}

	if _, err := ph.service.UpdateProject(ctx, id, updateData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Member removed successfully")
}
