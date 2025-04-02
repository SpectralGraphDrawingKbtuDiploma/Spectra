package service

import (
	"backend/internal/models"
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

func (t *TaskService) GetTaskByStatus(status string) (*models.Task, error) {
	query := `
		SELECT id, file_id, nodes_count, edges_array, mapping_array, status, created_at, updated_at
		FROM tasks
		WHERE status = $1
		ORDER BY RANDOM()
		LIMIT 1
	`

	var task models.Task
	var edgesArray, mappingArray pq.Int64Array

	err := t.DB.QueryRow(query, status).Scan(
		&task.ID,
		&task.FileID,
		&task.NodesCount,
		&edgesArray,
		&mappingArray,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error querying task with status %s: %w", status, err)
	}

	task.EdgesArray = make([]int, len(edgesArray))
	for i, v := range edgesArray {
		task.EdgesArray[i] = int(v)
	}

	task.MappingArray = make([]int, len(mappingArray))
	for i, v := range mappingArray {
		task.MappingArray[i] = int(v)
	}

	return &task, nil
}

func (t *TaskService) UpdateTaskStatus(id int, status string) error {
	query := `UPDATE tasks SET status = $1 WHERE id = $2`

	result, err := t.DB.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no job found with ID: %d", id)
	}

	return nil
}

// CompleteTask updates the status of a task and optionally sets an error message
func (t *TaskService) CompleteTaskInTx(id int, status string, errorMsg *string, resURL *string, tx *sql.Tx) error {
	var query string
	var args []interface{}

	if errorMsg == nil {
		// If no error message is provided, just update the status
		query = `UPDATE tasks SET status = $1 WHERE id = $2`
		args = []interface{}{status, id}
	} else {
		// If error message is provided, update both status and error
		query = `UPDATE tasks SET status = $1, error = $2 WHERE id = $3`
		args = []interface{}{status, *errorMsg, id}
	}

	// Execute the update statement
	result, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	// Check if any row was affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	// If no rows affected, it means no job with the given ID exists
	if rowsAffected == 0 {
		return fmt.Errorf("no job found with ID: %d", id)
	}

	if resURL == nil {
		return nil
	}

	// If no error message is provided, just update the status
	query = `UPDATE tasks SET result_url = $1 WHERE id = $2`
	args = []interface{}{*resURL, id}

	// Execute the update statement
	result, err = tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	// Check if any row was affected
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	// If no rows affected, it means no job with the given ID exists
	if rowsAffected == 0 {
		return fmt.Errorf("no job found with ID: %d", id)
	}

	return nil
}
