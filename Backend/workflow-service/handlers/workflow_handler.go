package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"workflow/models"
	"workflow/services"

	"github.com/gorilla/mux"
)

type WorkflowHandler struct {
	service *services.WorkflowService
}

func NewWorkflowHandler(service *services.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{service: service}
}

func (h *WorkflowHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if task.Title == "" || task.Project == "" {
		http.Error(w, "Title and Project are required", http.StatusBadRequest)
		return
	}

	if task.Status == "" {
		task.Status = models.Pending
	}

	exists, err := h.service.TaskExists(task.ID)
	if err != nil {
		http.Error(w, "Failed to check task existence", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Task with the same ID already exists", http.StatusConflict)
		return
	}

	err = h.service.CreateTask(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task created successfully"})
}

func (h *WorkflowHandler) CreateDependency(w http.ResponseWriter, r *http.Request) {
	var req models.DependencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err := h.service.CreateDependency(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Dependency created successfully"})
}

func (h *WorkflowHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.GetTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func (h *WorkflowHandler) GetTasksWithDependencies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	tasks, err := h.service.GetTasksWithDependencies(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func (h *WorkflowHandler) GetAllDependencies(w http.ResponseWriter, r *http.Request) {
	dependencies, err := h.service.GetAllDependencies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dependencies)
}

func (h *WorkflowHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	task, err := h.service.GetTask(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *WorkflowHandler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	fmt.Println("USAO OVDEEEEE", taskID)
	task, err := h.service.GetTask(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Pozvao servis za get tasks", task.Status)
	var updateStatus models.Status
	switch task.Status {
	case models.Pending:
		updateStatus = models.InProgress
	case models.InProgress:
		updateStatus = models.Finished
	case models.Finished:
		updateStatus = models.InProgress
	default:
		http.Error(w, "Invalid task status transition", http.StatusBadRequest)
		return
	}

	task.Status = updateStatus
	fmt.Println("Promenio status", task.Status)

	err = h.service.UpdateTask(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Pozvao servis za update task", task.Status)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task status updated successfully"})
}
