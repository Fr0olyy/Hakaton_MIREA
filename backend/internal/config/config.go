package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr     string
	DatabaseURL  string
	StorageDir   string
	MLServiceURL string
	MLTimeout    time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:     getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/hakaton?sslmode=disable"),
		StorageDir:   getEnv("STORAGE_DIR", "storage"),
		MLServiceURL: getEnv("ML_SERVICE_URL", "http://ml-service:8000"),
		MLTimeout:    time.Duration(getEnvInt("ML_TIMEOUT_SECONDS", 15)) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
