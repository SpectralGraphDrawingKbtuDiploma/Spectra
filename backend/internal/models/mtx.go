package models

import (
	"time"
)

// MtxFile represents an MTX file in the database
type MtxFile struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	Content    string    `json:"content,omitempty"`
	Dimensions string    `json:"dimensions"`
	CreatedAt  time.Time `json:"created_at"`
}

// MtxFileList is used for list responses where content is omitted
type MtxFileList struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	Dimensions string    `json:"dimensions"`
	CreatedAt  time.Time `json:"created_at"`
}
