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

	task.Blocked = false // Default status for a new task
	if err := h.service.CreateTask(&task); err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *WorkflowHandler) CreateDependency(w http.ResponseWriter, r *http.Request) {
	var dep models.Dependency
	if err := json.NewDecoder(r.Body).Decode(&dep); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateDependency(&dep); err != nil {
		http.Error(w, "Failed to create dependency", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
