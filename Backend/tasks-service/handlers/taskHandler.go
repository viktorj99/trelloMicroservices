package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"tasks-service/client"
	"tasks-service/model"
	"tasks-service/services"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskHandler struct {
	service       *services.TaskService
	logger        *log.Logger
	projectClient *client.ProjectClient
}

func NewTaskHandler(service *services.TaskService, logger *log.Logger, projectClient *client.ProjectClient) *TaskHandler {
	return &TaskHandler{
		service:       service,
		logger:        logger,
		projectClient: projectClient,
	}
}

func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tasks, err := h.service.GetAllTasks(ctx)
	if err != nil {
		h.logger.Println("Error fetching tasks:", err)
		http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetTaskById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	ctx := r.Context()
	task, err := h.service.GetTaskById(ctx, id)
	if err != nil {
		h.logger.Println("Error fetching task by ID:", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task model.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.logger.Println("Error decoding task:", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	createdTask, err := h.service.CreateTask(ctx, &task)
	if err != nil {
		h.logger.Println("Error creating task:", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTask)
}

func (h *TaskHandler) AssignMemberToTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskID"]
	memberID := vars["memberID"]

	ctx := r.Context()

	// Convert taskID and memberID to ObjectID
	taskObjectID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		h.logger.Println("Invalid task ID format:", err)
		http.Error(w, "Invalid task ID format", http.StatusBadRequest)
		return
	}

	memberObjectID, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		h.logger.Println("Invalid member ID format:", err)
		http.Error(w, "Invalid member ID format", http.StatusBadRequest)
		return
	}

	// Convert taskObjectID to string for GetTaskById call
	taskIDStr := taskObjectID.Hex()
	task, err := h.service.GetTaskById(ctx, taskIDStr)
	if err != nil {
		h.logger.Println("Error fetching task:", err)
		http.Error(w, "Failed to fetch task", http.StatusInternalServerError)
		return
	}

	// Get projectID from task
	projectID := task.Project.Hex()

	// Check if the member is part of the project
	resp, err := h.projectClient.CheckMemberInProject(ctx, projectID, memberID)
	if err != nil {
		h.logger.Println("Error checking if member is part of project:", err)
		http.Error(w, "Failed to check member's project status", http.StatusInternalServerError)
		return
	}

	// If the member is not part of the project, return an error
	if !resp.GetValue() {
		h.logger.Println("Member is not part of the project")
		http.Error(w, "Member is not part of the project", http.StatusBadRequest)
		return
	}

	// Proceed with assigning the member to the task
	updateData := bson.M{
		"$set": bson.M{
			"member": memberObjectID,
		},
	}

	err = h.service.UpdateTask(ctx, taskObjectID, updateData)
	if err != nil {
		h.logger.Println("Error assigning member to task:", err)
		http.Error(w, "Failed to assign member to task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Member assigned to task successfully"))
}

func (h *TaskHandler) GetTasksByProjectId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectId := vars["projectId"]

	ctx := r.Context()
	tasks, err := h.service.GetTasksByProjectId(ctx, projectId)
	if err != nil {
		h.logger.Println("Error fetching tasks by project ID:", err)
		http.Error(w, "Failed to retrieve tasks by project ID", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) ToggleTaskStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskID"]
	memberID := vars["memberID"]

	taskObjectID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		h.logger.Println("Invalid task ID format:", err)
		http.Error(w, "Invalid task ID format", http.StatusBadRequest)
		return
	}

	memberObjectID, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		h.logger.Println("Invalid member ID format:", err)
		http.Error(w, "Invalid member ID format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	task, err := h.service.GetTaskById(ctx, taskID)
	if err != nil {
		h.logger.Println("Error fetching task:", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	var updateStatus model.Status
	switch task.Status {
	case model.Pending:
		updateStatus = model.InProgress
	case model.InProgress:
		updateStatus = model.Finished
	case model.Finished:
		updateStatus = model.InProgress
	default:
		http.Error(w, "Invalid task status transition", http.StatusBadRequest)
		return
	}

	updateData := bson.M{
		"$set": bson.M{
			"member": memberObjectID,
			"status": updateStatus,
		},
	}

	err = h.service.UpdateTask(ctx, taskObjectID, updateData)
	if err != nil {
		h.logger.Println("Error updating task status:", err)
		http.Error(w, "Failed to update task status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task status updated successfully"))
}

func (h *TaskHandler) RemoveMemberFromTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskID"]

	ctx := r.Context()
	taskObjectID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		h.logger.Println("Invalid task ID format:", err)
		http.Error(w, "Invalid task ID format", http.StatusBadRequest)
		return
	}

	task, err := h.service.GetTaskById(ctx, taskID)
	if err != nil {
		h.logger.Println("Error fetching task:", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if task.Status == model.Finished {
		http.Error(w, "Cannot remove member from a finished task", http.StatusBadRequest)
		return
	}

	if task.Status == model.InProgress {
		updateData := bson.M{
			"$set":   bson.M{"status": model.Pending},
			"$unset": bson.M{"member": ""},
		}
		err = h.service.UpdateTask(ctx, taskObjectID, updateData)
		if err != nil {
			h.logger.Println("Error removing member and updating task status:", err)
			http.Error(w, "Failed to remove member and update task status", http.StatusInternalServerError)
			return
		}
	} else {
		updateData := bson.M{
			"$unset": bson.M{"member": ""},
		}
		err = h.service.UpdateTask(ctx, taskObjectID, updateData)
		if err != nil {
			h.logger.Println("Error removing member from task:", err)
			http.Error(w, "Failed to remove member from task", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Member removed from task successfully"))
}
