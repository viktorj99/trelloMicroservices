package handlers

import (
	"encoding/json"
	"net/http"
	"workflow/models"
	"workflow/services"
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

	// Proveri da li već postoji task sa istim ID-jem
	exists, err := h.service.TaskExists(task.ID)
	if err != nil {
		http.Error(w, "Failed to check task existence", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Task with the same ID already exists", http.StatusConflict)
		return
	}

	// Kreiraj task
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
	tasks, err := h.service.GetTasksWithDependencies()
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
