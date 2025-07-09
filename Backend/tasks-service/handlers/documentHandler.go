package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"tasks-service/services"

	"github.com/gorilla/mux"
)

type DocumentHandler struct {
	service *services.DocumentService
	logger  *log.Logger
}

func NewDocumentHandler(service *services.DocumentService, logger *log.Logger) *DocumentHandler {
	return &DocumentHandler{
		service: service,
		logger:  logger,
	}
}

// UploadDocument handles file upload to a task's document folder
func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskId"]

	// Parse multipart form (max 10 MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File not provided: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	userID := r.FormValue("userID")
	fileName := header.Filename

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Upload via service
	ctx := r.Context()
	document, err := h.service.UploadDocument(ctx, taskID, userID, fileName, fileBytes)
	if err != nil {
		http.Error(w, "Error uploading document: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(document)
}

// GetDocumentsHandler retrieves documents for a task
func (h *DocumentHandler) GetDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskId"]

	ctx := r.Context()
	documents, err := h.service.GetDocumentsByTaskID(ctx, taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching documents: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

// DownloadDocument handles downloading a file from HDFS
func (h *DocumentHandler) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskId"]
	docName := vars["docName"]

	filePath := fmt.Sprintf("/tasks/%s/%s", taskID, docName)

	content, err := h.service.DownloadDocument(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error downloading document: %v", err), http.StatusInternalServerError)
		return
	}

	// Detect mime type from extension (you can improve this detection)
	ext := strings.ToLower(filepath.Ext(docName))
	mimeType := "application/octet-stream"
	if ext == ".pdf" {
		mimeType = "application/pdf"
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", docName))
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))

	w.Write(content)
}
