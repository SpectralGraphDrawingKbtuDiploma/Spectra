package models

import (
	"time"
)

type Job struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	Content    string    `json:"content,omitempty"`
	Dimensions string    `json:"dimensions"`
	CreatedAt  time.Time `json:"created_at"`
	Status     string    `json:"status"`
	Error      *string   `json:"error,omitempty"`
	ResUrl     *string   `json:"res_url,omitempty"`
}

type JobList struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	Dimensions string    `json:"dimensions"`
	CreatedAt  time.Time `json:"created_at"`
}
