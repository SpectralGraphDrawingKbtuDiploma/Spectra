package models

import "time"

type Task struct {
	ID           int
	FileID       int
	NodesCount   int
	EdgesArray   []int
	MappingArray []int
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
