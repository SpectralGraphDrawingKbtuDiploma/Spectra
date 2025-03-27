package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"backend/internal/service"
	"github.com/gorilla/mux"
)

// JobsHandler handles HTTP requests for MTX files
type JobsHandler struct {
	Service *service.JobService
}

// NewJobsHandler creates a new MTX handler
func NewJobsHandler(mtxService *service.JobService) *JobsHandler {
	return &JobsHandler{
		Service: mtxService,
	}
}

// UploadJob handles MTX file uploads
func (h *JobsHandler) UploadJob(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("mtxfile")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save to database
	id, err := h.Service.SaveJob(header.Filename, string(content))
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "File uploaded successfully",
		"id":      id,
	})
}

// GetJob returns an MTX file by ID
func (h *JobsHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL params
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Fetch file from database
	file, err := h.Service.GetJob(id)
	if err != nil {
		if err.Error() == "file not found" {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Return file data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}

// ListJobs returns list of all MTX files
func (h *JobsHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	files, err := h.Service.ListJobs(false)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// DownloadJob provides raw MTX file content for download
func (h *JobsHandler) DownloadJob(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL params
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Fetch file from database
	file, err := h.Service.GetJob(id)
	if err != nil {
		if err.Error() == "file not found" {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Filename))
	w.Write([]byte(file.Content))
}
