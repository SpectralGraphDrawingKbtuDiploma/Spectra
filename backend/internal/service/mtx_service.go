package service

import (
	"database/sql"
	"errors"

	"backend/internal/models"
	"backend/pkg/mtxparser"
)

// MtxService provides business logic for MTX file operations
type MtxService struct {
	DB *sql.DB
}

// NewMtxService creates a new MTX service
func NewMtxService(db *sql.DB) *MtxService {
	return &MtxService{DB: db}
}

// SaveMtxFile saves MTX file content to database
func (s *MtxService) SaveMtxFile(filename string, content string) (int, error) {
	// Parse dimensions from content
	dimensions := mtxparser.ParseDimensions(content)

	var id int
	err := s.DB.QueryRow(
		"INSERT INTO mtx_files (filename, content, dimensions) VALUES ($1, $2, $3) RETURNING id",
		filename, content, dimensions,
	).Scan(&id)

	return id, err
}

// GetMtxFile retrieves MTX file from database by ID
func (s *MtxService) GetMtxFile(id int) (models.MtxFile, error) {
	var file models.MtxFile
	err := s.DB.QueryRow(
		"SELECT id, filename, content, dimensions, created_at FROM mtx_files WHERE id = $1",
		id,
	).Scan(&file.ID, &file.Filename, &file.Content, &file.Dimensions, &file.CreatedAt)

	if err == sql.ErrNoRows {
		return file, errors.New("file not found")
	}

	return file, err
}

// ListMtxFiles returns a list of all MTX files
func (s *MtxService) ListMtxFiles() ([]models.MtxFileList, error) {
	rows, err := s.DB.Query(
		"SELECT id, filename, dimensions, created_at FROM mtx_files ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.MtxFileList
	for rows.Next() {
		var file models.MtxFileList
		if err := rows.Scan(&file.ID, &file.Filename, &file.Dimensions, &file.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}
