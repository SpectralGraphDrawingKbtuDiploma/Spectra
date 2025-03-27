package service

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq" // Import for PostgreSQL array support
)

type TaskService struct {
	DB *sql.DB
}

func NewTaskService(db *sql.DB) *TaskService {
	return &TaskService{
		DB: db,
	}
}

// CreateTask creates a new task in the database with the given node count, mapping, and file ID
// Note: The function signature doesn't include edges, so we'll need to add a parameter for it
func (t *TaskService) CreateTaskInTx(nodesCount int, mapping []int, fileID int, edges []int, tx *sql.Tx) error {
	query := `
		INSERT INTO tasks (file_id, nodes_count, edges_array, mapping_array, status)
		VALUES ($1, $2, $3, $4, 'created')
		RETURNING id
	`

	var taskID int
	err := tx.QueryRow(
		query,
		fileID,
		nodesCount,        // Store as integer
		pq.Array(edges),   // Store as PostgreSQL array
		pq.Array(mapping), // Store as PostgreSQL array
	).Scan(&taskID)

	if err != nil {
		return fmt.Errorf("error inserting task: %w", err)
	}

	return nil
}
