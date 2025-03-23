package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading it, using environment variables")
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "postgres"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "mtxdb"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
