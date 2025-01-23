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
	userClient *client.UserClient
}

// NewProjectHandler creates a new ProjectHandler instance
func NewProjectHandler(service *services.ProjectService, taskClient *client.TaskClient, userClient *client.UserClient) *ProjectHandler {
	return &ProjectHandler{
		service:    service,
		taskClient: taskClient,
		userClient: userClient,
	}
}

func (ph *ProjectHandler) GetAllProjects(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "GetAllProjectsHandler")
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
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "GetProjectByIdHandler")
	defer span.End()

	userId := r.Header.Get("user_id")

	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		span.RecordError(err)
		LogEvent("2001", Error, "Invalid ID format", userId, id.Hex())
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	project, err := ph.service.GetProjectById(ctx, id)
	if err != nil {
		span.RecordError(err)
		LogEvent("2002", Error, "Project not found", userId, id.Hex())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetAttributes(attribute.String("project.id", id.Hex()))
	LogEvent("2000", Success, "Successfully fetched project by ID", userId, id.Hex())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (ph *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "CreateProjectHandler")
	defer span.End()

	var project model.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		span.RecordError(err)
		LogEvent("1001", Error, "Invalid request payload", project.Manager.ID.Hex(), "")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	LogEvent("1002", Success, "Project creation initiated", project.Manager.ID.Hex(), "")

	exists, err := ph.userClient.CheckIfUserExists(ctx, project.Manager.ID.Hex())
	if err != nil {
		LogEvent("1003", Error, "Error checking manager existence", project.Manager.ID.Hex(), "")
		http.Error(w, "Error checking manager existence", http.StatusInternalServerError)
		return
	}
	if !exists {
		LogEvent("1004", Error, "Manager does not exist", project.Manager.ID.Hex(), "")
		http.Error(w, "Manager does not exist", http.StatusBadRequest)
		return
	}

	for _, member := range project.Members {
		memberExists, err := ph.userClient.CheckIfUserExists(ctx, member.ID.Hex())
		if err != nil {
			LogEvent("1005", Error, "Error checking member existence", member.ID.Hex(), "")
			http.Error(w, "Error checking member existence", http.StatusInternalServerError)
			return
		}
		if !memberExists {
			LogEvent("1006", Error, "One or more members do not exist", member.ID.Hex(), "")
			http.Error(w, "One or more members do not exist", http.StatusBadRequest)
			return
		}
	}

	sanitizer := bluemonday.StrictPolicy()
	project.Name = sanitizer.Sanitize(project.Name)

	existingProject, err := ph.service.GetProjectByName(ctx, project.Name)
	if err != nil {
		span.RecordError(err)
		LogEvent("1007", Error, "Error checking project name", project.Manager.ID.Hex(), "")
		http.Error(w, "Error checking project name", http.StatusInternalServerError)
		return
	}
	if existingProject != nil {
		LogEvent("1008", Error, "A project with this name already exists", project.Manager.ID.Hex(), "")
		http.Error(w, "A project with this name already exists", http.StatusBadRequest)
		return
	}

	if project.ExpectedEndDate.IsZero() || project.ExpectedEndDate.Time.Before(time.Now()) {
		LogEvent("1009", Error, "Invalid expected end date", project.Manager.ID.Hex(), "")
		http.Error(w, "Expected end date cannot be empty or in the past", http.StatusBadRequest)
		return
	}

	if project.MinMembers <= 0 || project.MaxMembers < project.MinMembers {
		LogEvent("1010", Error, "Invalid member limits", project.Manager.ID.Hex(), "")
		http.Error(w, "Invalid member limits", http.StatusBadRequest)
		return
	}

	if len(project.Members) < project.MinMembers || len(project.Members) > project.MaxMembers {
		LogEvent("1011", Error, "Invalid number of members", project.Manager.ID.Hex(), "")
		http.Error(w, "Invalid number of members", http.StatusBadRequest)
		return
	}

	result, err := ph.service.CreateProject(ctx, &project)
	if err != nil {
		span.RecordError(err)
		LogEvent("1012", Error, "Failed to create project", project.Manager.ID.Hex(), "")
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	var projectID string
	projectID = result.InsertedID.(primitive.ObjectID).Hex()

	span.SetAttributes(attribute.String("project.name", project.Name))
	LogEvent("1000", Success, "Project created successfully", project.Manager.ID.Hex(), projectID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (ph *ProjectHandler) AddMemberToProject(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "AddMemberToProjectHandler")
	defer span.End()

	TokenUserId := r.Header.Get("user_id")

	vars := mux.Vars(r)
	projectID := vars["id"]

	id, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		span.RecordError(err)
		LogEvent("3001", Error, "Failed to Add Member - Invalid Project ID", TokenUserId, projectID)
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var member model.User
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		span.RecordError(err)
		LogEvent("3002", Error, "Failed to Add Member - Invalid member data", TokenUserId, projectID)
		http.Error(w, "Invalid member data", http.StatusBadRequest)
		return
	}

	exists, err := ph.userClient.CheckIfUserExists(ctx, member.ID.Hex())
	if err != nil {
		LogEvent("3003", Error, "Failed to Add Member - Error checking user existence", TokenUserId, projectID)
		http.Error(w, "Error checking user existence", http.StatusInternalServerError)
		return
	}
	if !exists {
		LogEvent("3004", Error, "Failed to Add Member - User does not exist", TokenUserId, projectID)
		http.Error(w, "User does not exist", http.StatusBadRequest)
		return
	}

	project, err := ph.service.GetProjectById(ctx, id)
	if err != nil {
		span.RecordError(err)
		LogEvent("3005", Error, "Failed to Add Member - Project not found", TokenUserId, projectID)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	for _, m := range project.Members {
		if m.ID == member.ID {
			LogEvent("3006", Error, "Failed to Add Member - Member: "+member.ID.Hex()+" already in project", TokenUserId, projectID)
			http.Error(w, "Member already in project", http.StatusBadRequest)
			return
		}
	}

	if len(project.Members)+1 > project.MaxMembers {
		LogEvent("3007", Error, "Failed to Add Member: "+member.ID.Hex()+" - Exceeds maximum limit", TokenUserId, projectID)
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
		LogEvent("3008", Error, "Failed to Add Member: Error fetching unassigned tasks", TokenUserId, projectID)
		http.Error(w, "Error fetching unassigned tasks", http.StatusInternalServerError)
		return
	}

	if len(resp.Tasks) > 0 {
		updateData := bson.M{
			"$push": bson.M{"members": member},
		}

		if _, err := ph.service.UpdateProject(ctx, id, updateData); err != nil {
			span.RecordError(err)
			LogEvent("3009", Error, "Internal Server Error", TokenUserId, projectID)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		span.SetAttributes(attribute.String("project.id", id.Hex()), attribute.String("member.id", member.ID.Hex()))
		w.WriteHeader(http.StatusOK)
		LogEvent("3000", Success, "Member: "+member.ID.Hex()+" added successfully", TokenUserId, projectID)
		json.NewEncoder(w).Encode("Member added successfully")
	} else {
		LogEvent("3010", Error, "No unassigned tasks found in the project", TokenUserId, projectID)
		http.Error(w, "No unassigned tasks found in the project", http.StatusConflict)
	}
}

func (ph *ProjectHandler) RemoveMemberFromProject(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "RemoveMemberFromProjectHandler")
	defer span.End()

	userId := r.Header.Get("user_id")

	vars := mux.Vars(r)
	projectID := vars["id"]

	id, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		span.RecordError(err)
		LogEvent("4001", Error, "Failed to Remove Member - Invalid Project ID", userId, projectID)
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var member model.User
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		span.RecordError(err)
		LogEvent("4002", Error, "Failed to Remove Member - Invalid member data", userId, projectID)
		http.Error(w, "Invalid member data", http.StatusBadRequest)
		return
	}

	project, err := ph.service.GetProjectById(ctx, id)
	if err != nil {
		span.RecordError(err)
		LogEvent("4003", Error, "Failed to Remove Member - Project not found", userId, projectID)
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
		LogEvent("4004", Error, "Failed to Remove Member - Member: "+member.ID.Hex()+" not found in project", userId, projectID)
		http.Error(w, "Member not found in project", http.StatusBadRequest)
		return
	}

	if len(project.Members)-1 < project.MinMembers {
		err := fmt.Errorf("removing member would fall below minimum members for project %s", projectID)
		span.RecordError(err)
		LogEvent("4005", Error, "Failed to Remove Member: "+member.ID.Hex()+" - nfalls below minimum limit", userId, projectID)
		http.Error(w, "Cannot remove member: falls below minimum limit", http.StatusBadRequest)
		return
	}

	taskReq := &taskpb.MemberRequest{MemberId: member.ID.Hex()}
	grpcSpanCtx, grpcSpan := otel.Tracer("projects-service").Start(ctx, "CheckMemberTasksInProgress")
	resp, err := ph.taskClient.CheckMemberTasksInProgress(grpcSpanCtx, taskReq)
	grpcSpan.End()

	if err != nil {
		grpcSpan.RecordError(err)
		LogEvent("4006", Error, "Error checking member tasks", userId, projectID)
		http.Error(w, "Error checking member tasks", http.StatusInternalServerError)
		return
	}

	if resp.GetValue() {
		err := fmt.Errorf("member %s cannot be removed because they are assigned to a task", member.ID.Hex())
		span.RecordError(err)
		LogEvent("4007", Error, "Failed to Remove Member: "+member.ID.Hex()+" - cannot be removed because they are assigned to a task", userId, projectID)
		http.Error(w, "Cannot remove member that is assigned to a task.", http.StatusBadRequest)
		return
	}

	updateData := bson.M{
		"$pull": bson.M{"members": bson.M{"_id": member.ID}},
	}

	if _, err := ph.service.UpdateProject(ctx, id, updateData); err != nil {
		span.RecordError(err)
		LogEvent("4008", Error, "Failed to Remove Member - Internal Server Error", userId, projectID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(
		attribute.String("project.id", projectID),
		attribute.String("member.id", member.ID.Hex()),
		attribute.String("action", "member_removed"),
	)

	w.WriteHeader(http.StatusOK)
	LogEvent("4000", Success, "Member: "+member.ID.Hex()+" removed successfully", userId, projectID)
	json.NewEncoder(w).Encode("Member removed successfully")
}

func (handler *ProjectHandler) FindProjectsByUserID(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "FindProjectsByUserIDHandler")
	defer span.End()

	vars := mux.Vars(r)
	userIDStr := vars["id"]
	if userIDStr == "" {
		span.AddEvent("Missing userID in request")
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("userID", userIDStr))

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		span.RecordError(err)
		span.AddEvent("Invalid userID format")
		http.Error(w, "Invalid userID format", http.StatusBadRequest)
		return
	}

	projects, err := handler.service.GetProjectsByUserID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.AddEvent("Error fetching projects for user ID")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("projects.count", len(projects)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		span.RecordError(err)
		span.AddEvent("Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (handler *ProjectHandler) FindProjectsByManagerID(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "FindProjectsByManagerIDHandler")
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
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "DeleteProjectHandler")
	defer span.End()

	userId := r.Header.Get("user_id")

	vars := mux.Vars(r)
	idParam := vars["id"]

	projectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		span.RecordError(err)
		LogEvent("5001", Error, "Failed to Delete Project - Invalid Project ID", userId, projectID.Hex())
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	err = ph.service.DeleteProjectWithTasks(ctx, projectID)
	if err != nil {
		span.RecordError(err)
		LogEvent("5002", Error, "Failed to Delete Project - Internal Server Error", userId, projectID.Hex())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.String("project.id", projectID.Hex()))
	w.WriteHeader(http.StatusAccepted)
	LogEvent("5000", Success, "Project: "+projectID.Hex()+" deleted successfully", userId, projectID.Hex())
	json.NewEncoder(w).Encode("Delete request accepted")
}

func (ph *ProjectHandler) FinishProject(w http.ResponseWriter, r *http.Request){
	ctx, span := otel.Tracer("project-service").Start(r.Context(), "FinishedProjectHandler")
	defer span.End()

	vars := mux.Vars(r)
	idParam := vars["id"]

	id, err := primitive.ObjectIDFromHex(idParam)
	if err!= nil {
		span.RecordError(err)
        http.Error(w, "Invalid project ID", http.StatusBadRequest)
        return
	}

	result, err := ph.service.FinishProject(ctx, id)
	if err!= nil {
        span.RecordError(err)
        http.Error(w, "Failed to finish project: "+err.Error(), http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
