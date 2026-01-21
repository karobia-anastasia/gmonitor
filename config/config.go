package config

import (
	"log"
	"os"
	"time"
)

// Config holds all configuration settings for the application
type Config struct {
	GitHubToken   string
	DatabaseURL   string
	PollInterval  time.Duration
	PORT          string
	RedisHost     string
	RedisPassword string
}

// LoadConfig initializes the configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		GitHubToken:   getEnv("GITHUB_TOKEN", ""),
		DatabaseURL:   getEnv("DB_DSN", "gmonitor.db"),
		PollInterval:  getEnvAsDuration("POLL_INTERVAL", time.Minute), // Default: 1 Minute
		PORT:          getEnv("SERVER_PORT", "8000"),
		RedisHost:     getEnv("REDIS_HOST", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
	}
}

// getEnv retrieves a string environment variable or uses a default value
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// getEnvAsDuration retrieves an environment variable as a time.Duration or uses a default
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("Invalid duration format for %s: %s, using default: %v", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}
