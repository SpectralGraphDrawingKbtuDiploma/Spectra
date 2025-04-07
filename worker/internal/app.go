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
	path := fmt.Sprintf("/var/worker/graph-%s", *graph.ID)
	logFile, _ := os.Create(filepath.Join(path, "log.txt"))
	defer logFile.Close()
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		_, _ = logFile.WriteString(fmt.Sprintf("Failed to create directory: %s\n", err))
		return
	}
	filePath := filepath.Join(path, "example.mtx")
	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		_, _ = logFile.WriteString(fmt.Sprintf("Failed to create file: %s\n", err))
		return
	}
	_, _ = file.WriteString(*graph.Content)
	cmd := exec.Command("sh", "draw.sh", fmt.Sprintf("%s/graph.txt", path), path, *graph.ID)
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
	if graph.Content != nil {
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
	path := fmt.Sprintf("/var/worker/graph-%s", *graph.ID)
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
			fmt.Println(entry.Name(), "$$$")
			filePath := filepath.Join(path, "result.txt")
			fmt.Println(filePath)
			res.Status = "completed"
			// Read result.txt file
			if resultContent, err := os.ReadFile(filePath); err == nil {
				// Save content to Result field
				content := string(resultContent)
				res.Result = &content
			} else {
				// If there's an error reading result.txt, report it
				errMsg := fmt.Sprintf("Failed to read result file: %v", err)
				res.Err = &errMsg
			}

			_ = encoder.Encode(res)
			return
		}
		if entry.Name() == "error.txt" {
			res.Status = "completed"
			errFilePath := filepath.Join(path, "error.txt")
			if errFileContent, err := os.ReadFile(errFilePath); err == nil && len(errFileContent) > 0 {
				errContent := string(errFileContent)
				res.Err = &errContent
			}
			_ = encoder.Encode(res)
			return
		}
	}
	err = encoder.Encode(res)
	if err != nil {
		app.logger.Error("err while encoding task status", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
