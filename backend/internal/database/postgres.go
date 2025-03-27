package database

import (
	"database/sql"
	"fmt"
	"log"

	"backend/internal/config"
	_ "github.com/lib/pq"
)

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

func InitSchema(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS jobs (
            id SERIAL PRIMARY KEY,
            filename VARCHAR(255) NOT NULL,
            content TEXT NOT NULL,
            dimensions VARCHAR(50),
            scheduled BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
		CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			file_id INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
			nodes_count INTEGER NOT NULL,
			edges_array INTEGER[],
			mapping_array INTEGER[],
			status VARCHAR(50) DEFAULT 'created',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
    `)

	return err
}
