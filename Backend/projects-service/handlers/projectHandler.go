package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	taskpb "pb/taskpb"
	"projects-service/client"
	"projects-service/model"
	"projects-service/services"
	"time"

	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type ProjectHandler struct {
	service    *services.ProjectService
	taskClient *client.TaskClient
}

// NewProjectHandler creates a new ProjectHandler instance
func NewProjectHandler(service *services.ProjectService, taskClient *client.TaskClient) *ProjectHandler {
	return &ProjectHandler{
		service:    service,
		taskClient: taskClient,
	}
}

func (ph *ProjectHandler) GetAllProjects(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "GetAllProjects")
	defer span.End()

	projects, err := ph.service.GetAllProjects(ctx)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("projects.count", len(projects)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (ph *ProjectHandler) GetProjectById(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "GetProjectById")
	defer span.End()

	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	project, err := ph.service.GetProjectById(ctx, id)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetAttributes(attribute.String("project.id", id.Hex()))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (ph *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "CreateProject")
	defer span.End()

	var project model.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	sanitizer := bluemonday.StrictPolicy()
	project.Name = sanitizer.Sanitize(project.Name)

	existingProject, err := ph.service.GetProjectByName(ctx, project.Name)
	if err != nil {
		span.RecordError(err)
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

	if project.MinMembers <= 0 || project.MaxMembers < project.MinMembers {
		http.Error(w, "Invalid member limits", http.StatusBadRequest)
		return
	}

	if len(project.Members) < project.MinMembers || len(project.Members) > project.MaxMembers {
		http.Error(w, "Invalid number of members", http.StatusBadRequest)
		return
	}

	result, err := ph.service.CreateProject(ctx, &project)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.String("project.name", project.Name))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (ph *ProjectHandler) AddMemberToProject(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "AddMemberToProject")
	defer span.End()

	vars := mux.Vars(r)
	projectID := vars["id"]

	id, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var member model.User
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid member data", http.StatusBadRequest)
		return
	}

	project, err := ph.service.GetProjectById(ctx, id)
	if err != nil {
		span.RecordError(err)
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

	// Call gRPC service to fetch unassigned tasks
	req := &taskpb.ProjectRequest{
		ProjectId: projectID,
	}
	resp, err := ph.taskClient.GetUnassignedTasks(ctx, req)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Error fetching unassigned tasks", http.StatusInternalServerError)
		return
	}

	if len(resp.Tasks) > 0 {
		updateData := bson.M{
			"$push": bson.M{"members": member},
		}

		if _, err := ph.service.UpdateProject(ctx, id, updateData); err != nil {
			span.RecordError(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		span.SetAttributes(attribute.String("project.id", id.Hex()), attribute.String("member.id", member.ID.Hex()))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("Member added successfully")
	} else {
		http.Error(w, "No unassigned tasks found in the project", http.StatusConflict)
	}
}

func (ph *ProjectHandler) RemoveMemberFromProject(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "RemoveMemberFromProject")
	defer span.End()

	vars := mux.Vars(r)
	projectID := vars["id"]

	id, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var member model.User
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid member data", http.StatusBadRequest)
		return
	}

	project, err := ph.service.GetProjectById(ctx, id)
	if err != nil {
		span.RecordError(err)
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
		err := fmt.Errorf("member %s not found in project %s", member.ID.Hex(), projectID)
		span.RecordError(err)
		http.Error(w, "Member not found in project", http.StatusBadRequest)
		return
	}

	if len(project.Members)-1 < project.MinMembers {
		err := fmt.Errorf("removing member would fall below minimum members for project %s", projectID)
		span.RecordError(err)
		http.Error(w, "Cannot remove member: falls below minimum limit", http.StatusBadRequest)
		return
	}

	taskReq := &taskpb.MemberRequest{MemberId: member.ID.Hex()}
	grpcSpanCtx, grpcSpan := otel.Tracer("projects-service").Start(ctx, "CheckMemberTasksInProgress")
	resp, err := ph.taskClient.CheckMemberTasksInProgress(grpcSpanCtx, taskReq)
	grpcSpan.End()

	if err != nil {
		grpcSpan.RecordError(err)
		http.Error(w, "Error checking member tasks", http.StatusInternalServerError)
		return
	}

	if resp.GetValue() {
		err := fmt.Errorf("member %s cannot be removed because they are assigned to a task", member.ID.Hex())
		span.RecordError(err)
		http.Error(w, "Cannot remove member that is assigned to a task.", http.StatusBadRequest)
		return
	}

	updateData := bson.M{
		"$pull": bson.M{"members": bson.M{"_id": member.ID}},
	}

	if _, err := ph.service.UpdateProject(ctx, id, updateData); err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(
		attribute.String("project.id", projectID),
		attribute.String("member.id", member.ID.Hex()),
		attribute.String("action", "member_removed"),
	)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Member removed successfully")
}

func (handler *ProjectHandler) FindProjectsByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	if userIDStr == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userID format", http.StatusBadRequest)
		return
	}
	projects, err := handler.service.GetProjectsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (handler *ProjectHandler) FindProjectsByManagerID(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "FindProjectsByManagerID")
	defer span.End()
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	if userIDStr == "" {
		err := fmt.Errorf("userID is required but not provided")
		span.RecordError(err)
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid userID format", http.StatusBadRequest)
		return
	}

	projects, err := handler.service.GetProjectsByManagerID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(
		attribute.String("manager.id", userID.Hex()),
		attribute.Int("projects.count", len(projects)),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (ph *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "DeleteProject")
	defer span.End()

	vars := mux.Vars(r)
	idParam := vars["id"]

	projectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	err = ph.service.DeleteProjectWithTasks(ctx, projectID)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.String("project.id", projectID.Hex()))
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode("Delete request accepted")
}
