package config

import (
	"os"
)

// Config holds all configuration for the server
type Config struct {
	ServerAddress     string
	CorsOrigins       string
	UseDynamoDB       bool
	DynamoDBTableName string
	JWKSUrl           string
	JWTIssuer         string
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		ServerAddress:     getEnv("SERVER_ADDRESS", ":8080"),
		CorsOrigins:       getEnv("CORS_ORIGINS", "*"),
		UseDynamoDB:       getEnvBool("USE_DYNAMODB", false),
		DynamoDBTableName: getEnv("DYNAMODB_TABLE_NAME", "messages"),
		JWKSUrl:           getEnv("JWKS_URL", ""),
		JWTIssuer:         getEnv("JWT_ISSUER", ""),
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

// getEnvBool gets an environment variable as a boolean or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}
