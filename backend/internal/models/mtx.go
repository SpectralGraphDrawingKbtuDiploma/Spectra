package models

import (
	"time"
)

// Job represents an MTX file in the database
type Job struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	Content    string    `json:"content,omitempty"`
	Dimensions string    `json:"dimensions"`
	CreatedAt  time.Time `json:"created_at"`
}

// JobList is used for list responses where content is omitted
type JobList struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	Dimensions string    `json:"dimensions"`
	CreatedAt  time.Time `json:"created_at"`
}
