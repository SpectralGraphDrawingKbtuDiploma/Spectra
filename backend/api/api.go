package api

import (
	"database/sql"
	"log"
	"net/http"

	"backend/internal/config"
	"backend/internal/handlers"
	"github.com/gorilla/mux"
)

// StartServer starts the HTTP server
func StartServer(cfg *config.Config, db *sql.DB) error {
	// Create handlers
	mtxHandler := handlers.NewMtxHandler(db)

	// Create router
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/mtx", mtxHandler.UploadMtx).Methods("POST")
	router.HandleFunc("/api/mtx", mtxHandler.ListMtx).Methods("GET")
	router.HandleFunc("/api/mtx/{id:[0-9]+}", mtxHandler.GetMtx).Methods("GET")
	router.HandleFunc("/api/mtx/{id:[0-9]+}/download", mtxHandler.DownloadMtx).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Start server
	log.Printf("Starting server on port %s", cfg.ServerPort)
	return http.ListenAndServe(":"+cfg.ServerPort, router)
}
