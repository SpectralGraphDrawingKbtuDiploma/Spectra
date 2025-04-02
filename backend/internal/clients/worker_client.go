package clients

import (
	"backend/internal/dto"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WorkerClient struct {
	httpClient *http.Client
	workerHost string
}

// NewWorkerClient creates a new worker client with the specified host
func NewWorkerClient(workerHost string) *WorkerClient {
	return &WorkerClient{
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		workerHost: workerHost,
	}
}

// Ping sends a ping request to the worker to check if it's available
func (c *WorkerClient) Ping(taskReq dto.TaskRequest) (*dto.TaskResponse, error) {
	// Construct the URL for the ping endpoint
	url := c.workerHost

	// Marshal the request to JSON
	reqBody, err := json.Marshal(taskReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create a new HTTP POST request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set appropriate headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("worker returned non-OK status code: %d", resp.StatusCode)
	}

	// Decode the response
	var taskResp dto.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &taskResp, nil
}

// SetTimeout allows configuring a custom timeout for the HTTP client
func (c *WorkerClient) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}
