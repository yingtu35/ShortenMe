package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	BaseURL string
	Port    string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		BaseURL: getEnvOrDefault("SHORTENME_URL", "http://localhost:8080"),
		Port:    getEnvOrDefault("PORT", "8080"),
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
