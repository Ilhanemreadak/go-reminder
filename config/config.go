package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	DatabasePath    string
	EncryptionKey   string
	ServerPort      string
	SessionDuration time.Duration
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		DatabasePath:    getEnv("DATABASE_PATH", "data/reminders.db"),
		EncryptionKey:   getEnv("ENCRYPTION_KEY", ""),
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		SessionDuration: getDurationEnv("SESSION_DURATION", 24*time.Hour),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDurationEnv retrieves a duration from environment variable (in hours) or returns default
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if hours, err := strconv.Atoi(value); err == nil {
			return time.Duration(hours) * time.Hour
		}
	}
	return defaultValue
}
