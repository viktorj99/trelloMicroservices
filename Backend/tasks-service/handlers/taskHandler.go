package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"tasks-service/model"
	"tasks-service/services"

	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskHandler struct {
	service *services.TaskService
	logger  *log.Logger
}

func NewTaskHandler(service *services.TaskService, logger *log.Logger) *TaskHandler {
	return &TaskHandler{
		service: service,
		logger:  logger,
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

	sanitizer := bluemonday.StrictPolicy()
	task.Title = sanitizer.Sanitize(task.Title)
	task.Description = sanitizer.Sanitize(task.Description)

	if len(task.Title) < 5 || len(task.Title) > 100 {
		http.Error(w, "Title must be between 5 and 100 characters", http.StatusBadRequest)
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

	updateData := bson.M{
		"member": memberObjectID,
		"status": model.InProgress,
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
	if task.Status == model.InProgress {
		updateStatus = model.Finished
	} else if task.Status == model.Finished {
		updateStatus = model.InProgress
	} else {
		http.Error(w, "Invalid task status", http.StatusBadRequest)
		return
	}

	updateData := bson.M{
		"member": memberObjectID,
		"status": updateStatus,
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
