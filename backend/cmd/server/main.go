package main

import (
	"log"

	"backend/api"
	"backend/internal/config"
	"backend/internal/database"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize database schema
	if err := database.InitSchema(db); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start API server
	if err := api.StartServer(cfg, db); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
