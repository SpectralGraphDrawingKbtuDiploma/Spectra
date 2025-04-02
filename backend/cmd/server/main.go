package main

import (
	"backend/internal/clients"
	"backend/internal/scheduler"
	"backend/internal/service"
	"log"
	"os"

	"backend/api"
	"backend/internal/config"
	"backend/internal/database"
)

func main() {
	// Load configuration
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
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

	jobCreatedCh := make(chan struct{}, 100)
	taskService := service.NewTaskService(db)
	mtxService := service.NewMtxService(db, jobCreatedCh)
	workerClient := clients.NewWorkerClient(cfg.WorkerHost)
	s := scheduler.NewScheduler(
		taskService,
		mtxService,
		logger,
		jobCreatedCh,
		db,
		workerClient,
	)
	s.Start()
	// Start API server
	if err := api.StartServer(cfg, mtxService); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
