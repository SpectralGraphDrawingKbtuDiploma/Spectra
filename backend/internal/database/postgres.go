package database

import (
	"database/sql"
	"fmt"
	"log"

	"backend/internal/config"
	_ "github.com/lib/pq"
)

// Connect establishes a connection to PostgreSQL
func Connect(cfg *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to PostgreSQL database")
	return db, nil
}

// InitSchema initializes database schema
func InitSchema(db *sql.DB) error {
	// Create mtx_files table if it doesn't exist
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS mtx_files (
            id SERIAL PRIMARY KEY,
            filename VARCHAR(255) NOT NULL,
            content TEXT NOT NULL,
            dimensions VARCHAR(50),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)

	return err
}
