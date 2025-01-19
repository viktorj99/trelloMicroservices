package handlers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

	workflowServiceURL := os.Getenv("WORKFLOW_SERVICE_URL")
	if workflowServiceURL == "" {
		workflowServiceURL = "https://workflow-service:8443"
	}

	h.logger.Printf("Attempting to call workflow service at: %s/workflow/tasks", workflowServiceURL)

	workflowTask := map[string]interface{}{
		"id":          createdTask.ID.Hex(),
		"title":       createdTask.Title,
		"description": createdTask.Description,
		"status":      string(createdTask.Status),
		"project":     createdTask.Project.Hex(),
		"member":      createdTask.Member.Hex(),
		"blocked":     createdTask.Blocked,
	}

	workflowJSON, err := json.Marshal(workflowTask)
	if err != nil {
		h.logger.Printf("Warning: Failed to marshal workflow task: %v", err)
	} else {
		fullURL := workflowServiceURL + "/workflow/tasks"
		h.logger.Printf("Making POST request to: %s", fullURL)
		
		workflowReq, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(workflowJSON))
		if err != nil {
			h.logger.Printf("Warning: Failed to create workflow request: %v", err)
		} else {
			workflowReq.Header.Set("Content-Type", "application/json")
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
			
			h.logger.Printf("Sending request to workflow service...")
			resp, err := client.Do(workflowReq)
			if err != nil {
				h.logger.Printf("Warning: Failed to sync with workflow service: %v", err)
			} else {
				defer resp.Body.Close()
				h.logger.Printf("Workflow service responded with status: %d", resp.StatusCode)
				
				body, _ := io.ReadAll(resp.Body)
				h.logger.Printf("Response body: %s", string(body))
				
				if resp.StatusCode != http.StatusCreated {
					h.logger.Printf("Warning: Workflow service responded with unexpected status: %d", resp.StatusCode)
				}
			}
		}
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
	projectID := vars["projectID"]

	ctx := r.Context()

	tasks, err := h.service.GetTasksWithDependencies(ctx, projectID)
	if err != nil {
		h.logger.Println("Error fetching tasks with dependencies:", err)
		http.Error(w, "Failed to retrieve tasks with dependencies", http.StatusInternalServerError)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		log.Fatalf("Invalid ObjectID: %v", err)
	}

	for _, task := range tasks {
		for _, dependency := range task.Dependencies {
			if dependency.ID == objectID && task.Status != model.Finished {
				fmt.Println("prviiiiii " + model.Finished)
				http.Error(w, "Task is blocked by task with title: " + task.Title, http.StatusBadRequest)
				return
			}
		}
	}

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

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("PUT", "https://workflow-service:8443/workflow/tasks/update-status/"+taskID, nil)
	if err != nil {
		h.logger.Println("Error creating PUT request:", err)
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		h.logger.Println("Error calling workflow service:", err)
		return
	}
	defer resp.Body.Close()

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

func (h *TaskHandler) GetTasksWithDependencies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["projectId"]

	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tasks, err := h.service.GetTasksWithDependencies(ctx, projectID)
	if err != nil {
		http.Error(w, "Failed to retrieve tasks with dependencies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

