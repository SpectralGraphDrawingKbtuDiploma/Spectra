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
}

type JobList struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	Dimensions string    `json:"dimensions"`
	CreatedAt  time.Time `json:"created_at"`
}
