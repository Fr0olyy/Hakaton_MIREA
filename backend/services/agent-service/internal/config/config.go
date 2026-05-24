package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr     string
	AgentBaseURL string
	AgentModel   string
	AgentTimeout time.Duration
	LogLevel     string
}

func Load() Config {
	return Config{
		HTTPAddr:     getEnv("HTTP_ADDR", ":8090"),
		AgentBaseURL: getEnv("AGENT_BASE_URL", "http://localhost:11434"),
		AgentModel:   getEnv("AGENT_MODEL", "gemma2:2b"),
		AgentTimeout: time.Duration(getEnvInt("AGENT_TIMEOUT_SECONDS", 120)) * time.Second,
		LogLevel:     getEnv("LOG_LEVEL", "info"),
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
