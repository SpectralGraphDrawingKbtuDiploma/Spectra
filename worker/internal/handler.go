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
)

type App struct {
	m                            *sync.Mutex
	processLimit                 int
	currentWorkingProcessesCount int
	logger                       zap.Logger
}

func (app *App) createJob(w http.ResponseWriter, graph GraphDTO) {
	app.m.Lock()
	if app.processLimit <= app.currentWorkingProcessesCount {
		app.m.Unlock()
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	app.m.Unlock()
	path := fmt.Sprintf("/var/worker/graph-%s", graph.ID)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		http.Error(w, "не удалось создать директорию для записи", http.StatusInternalServerError)
		return
	}
	filePath := filepath.Join(path, "graph.txt")
	file, err := os.Create(filePath)
	defer file.Close()
	for i := 0; i < len(graph.Edges); i += 2 {
		_, err = file.Write([]byte(fmt.Sprintf("%v %v", graph.Edges[i], graph.Edges[i+1])))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			_ = os.Remove(filePath)
			return
		}
	}
	cmd := exec.Command("ls -l")
	err = cmd.Start()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		_ = os.Remove(filePath)
	}
	encoder := json.NewEncoder(w)
	res := TaskStatus{
		ID:     *graph.ID,
		Status: "created",
	}
	err = encoder.Encode(res)
	if err != nil {
		app.logger.Error("err while encoding task status", zap.Error(err))
	}
}

func (app *App) PingHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var graph GraphDTO
	err := decoder.Decode(&graph)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if graph.Edges != nil {
		app.createJob(w, graph)
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
		}
	}
	err = encoder.Encode(res)
	if err != nil {
		app.logger.Error("err while encoding task status", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
