package config

import (
	"os"
)

// Config holds all configuration for the server
type Config struct {
	ServerAddress string
	CorsOrigins   string
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		CorsOrigins:   getEnv("CORS_ORIGINS", "*"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
