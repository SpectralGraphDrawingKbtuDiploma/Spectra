package service

import (
	"database/sql"
	"errors"
	"strings"

	"backend/internal/models"
	"backend/pkg/mtxparser"
)

type JobService struct {
	DB           *sql.DB
	jobCreatedCh chan struct{}
}

func NewMtxService(db *sql.DB, jobCreatedCh chan struct{}) *JobService {
	return &JobService{DB: db, jobCreatedCh: jobCreatedCh}
}

func (s *JobService) checkDimensions(dimension string) error {
	parts := strings.Split(dimension, ":")
	if len(parts) != 2 {
		return errors.New("invalid dimension")
	}
	if parts[0] != parts[1] {
		return errors.New("dimension mismatch")
	}
	return nil
}

func (s *JobService) SaveJob(filename string, content string) (int, error) {
	// Parse dimensions from content
	dimensions := mtxparser.ParseDimensions(content)

	var id int
	err := s.DB.QueryRow(
		"INSERT INTO jobs (filename, content, dimensions) VALUES ($1, $2, $3) RETURNING id",
		filename, content, dimensions,
	).Scan(&id)
	s.jobCreatedCh <- struct{}{}
	return id, err
}

func (s *JobService) GetJob(id int) (models.Job, error) {
	var file models.Job
	err := s.DB.QueryRow(
		"SELECT id, filename, content, dimensions, created_at, status FROM jobs WHERE id = $1",
		id,
	).Scan(&file.ID, &file.Filename, &file.Content, &file.Dimensions, &file.CreatedAt, &file.Status)

	if err == sql.ErrNoRows {
		return file, errors.New("file not found")
	}

	return file, err
}

func (s *JobService) ListJobs(status *string) ([]models.JobList, error) {
	var rows *sql.Rows
	var err error
	que := "SELECT id, filename, dimensions, created_at FROM jobs ORDER BY created_at DESC"
	if status != nil {
		que = "SELECT id, filename, dimensions, created_at FROM jobs WHERE status=$1 ORDER BY created_at DESC"
		rows, err = s.DB.Query(
			que,
			*status,
		)
	} else {
		rows, err = s.DB.Query(
			que,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.JobList
	for rows.Next() {
		var file models.JobList
		if err := rows.Scan(&file.ID, &file.Filename, &file.Dimensions, &file.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

func (s *JobService) ScheduleJobInTx(id int, tx *sql.Tx) error {
	que := "UPDATE jobs SET status='executing' WHERE id = $1"
	_, err := tx.Exec(que, id)
	return err
}
