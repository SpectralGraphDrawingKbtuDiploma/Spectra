package internal

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type App struct {
	m                            *sync.Mutex
	processLimit                 int
	currentWorkingProcessesCount int
	logger                       *zap.Logger
	isClosed                     atomic.Bool
	allDone                      chan struct{}
}

func NewApp(logger *zap.Logger) *App {
	return &App{
		processLimit:                 1,
		currentWorkingProcessesCount: 0,
		m:                            &sync.Mutex{},
		logger:                       logger,
	}
}

func (app *App) GracefulShutdown() {
	app.isClosed.Store(true)
	select {
	case <-app.allDone:
		fmt.Println("All processes have been shut down")
	case <-time.After(5 * time.Second):
		fmt.Println("Timeout waiting for all processes to terminate")
	}
}

func (app *App) createJob(graph GraphDTO) {
	app.m.Lock()
	app.currentWorkingProcessesCount++
	app.m.Unlock()
	defer func() {
		app.m.Lock()
		app.currentWorkingProcessesCount--
		if app.isClosed.Load() && app.currentWorkingProcessesCount == 0 {
			app.allDone <- struct{}{}
		}
		app.m.Unlock()
	}()
	path := fmt.Sprintf("/var/worker/graph-%s", graph.ID)
	logFile, _ := os.Create(filepath.Join(path, "log.txt"))
	defer logFile.Close()
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		_, _ = logFile.WriteString(fmt.Sprintf("Failed to create directory: %s\n", err))
		return
	}
	filePath := filepath.Join(path, "graph.txt")
	file, err := os.Create(filePath)
	defer file.Close()
	for i := 0; i < len(graph.Edges); i += 2 {
		_, err = file.Write([]byte(fmt.Sprintf("%v %v", graph.Edges[i], graph.Edges[i+1])))
		if err != nil {
			_, _ = logFile.WriteString(fmt.Sprintf("Failed to write to file: %v\n", err))
			return
		}
	}
	cmd := exec.Command("ls -l")
	err = cmd.Start()
	if err != nil {
		_, _ = logFile.WriteString(fmt.Sprintf("Failed to start command: %v\n", err))
		return
	}
}

func (app *App) PingHandler(w http.ResponseWriter, r *http.Request) {
	if app.isClosed.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var graph GraphDTO
	err := decoder.Decode(&graph)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if graph.Edges != nil {
		go app.createJob(graph)
		w.WriteHeader(http.StatusOK)
		res := TaskStatus{
			ID:     *graph.ID,
			Status: "created",
		}
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(res)
		return
	}
	path := fmt.Sprintf("/var/worker/graph-%s", graph.ID)
	entries, err := os.ReadDir(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	encoder := json.NewEncoder(w)
	res := TaskStatus{
		ID:     *graph.ID,
		Status: "processing",
	}
	for _, entry := range entries {
		if entry.Name() == "result.txt" {
			res.Status = "completed"
			res.Result = []float64{1.0, 2.0}
			break
		}
	}
	err = encoder.Encode(res)
	if err != nil {
		app.logger.Error("err while encoding task status", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
