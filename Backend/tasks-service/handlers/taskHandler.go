package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"tasks-service/model"
	"tasks-service/services"

	"github.com/gorilla/mux"
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
