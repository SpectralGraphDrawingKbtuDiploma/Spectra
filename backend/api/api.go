package api

import (
	"backend/internal/service"
	"log"
	"net/http"

	"backend/internal/config"
	"backend/internal/handlers"
	"github.com/gorilla/mux"
)

// StartServer starts the HTTP server
func StartServer(cfg *config.Config, service *service.JobService) error {
	// Create handlers
	mtxHandler := handlers.NewJobsHandler(service)

	// Create router
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/jobs", mtxHandler.UploadJob).Methods("POST")
	router.HandleFunc("/api/jobs", mtxHandler.ListJobs).Methods("GET")
	router.HandleFunc("/api/jobs/{id:[0-9]+}", mtxHandler.GetJob).Methods("GET")
	router.HandleFunc("/api/jbos/{id:[0-9]+}/download", mtxHandler.DownloadJob).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Start server
	log.Printf("Starting server on port %s", cfg.ServerPort)
	return http.ListenAndServe(":"+cfg.ServerPort, router)
}
