package config

import (
	"log"
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	// Server configuration
	ServerAddress string

	// CORS configuration
	CorsOrigins string

	// Environment
	Environment string

	// DynamoDB configuration
	UseDynamoDB       bool
	DynamoDBTableName string

	// Cognito configuration
	UserPoolID       string
	UserPoolClientID string
	CognitoRegion    string
}

// NewConfig creates a new configuration from environment variables
func NewConfig() *Config {
	// Get server address from environment or use default
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = ":8081" // Default to port 8081 to avoid conflict with msgsvc
	}

	// Get CORS origins from environment or use default
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*" // Default to allow all origins
	}

	// Get environment name
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "dev" // Default to dev environment
	}

	// DynamoDB configuration
	useDynamoDB := false
	useDynamoDBStr := os.Getenv("USE_DYNAMODB")
	if useDynamoDBStr != "" {
		var err error
		useDynamoDB, err = strconv.ParseBool(useDynamoDBStr)
		if err != nil {
			log.Printf("WARNING: Invalid USE_DYNAMODB value: %s, defaulting to false", useDynamoDBStr)
		}
	}

	dynamoDBTableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if dynamoDBTableName == "" {
		dynamoDBTableName = "users" // Default table name
	}

	// Cognito configuration
	userPoolID := os.Getenv("COGNITO_USER_POOL_ID")
	if userPoolID == "" {
		log.Println("WARNING: COGNITO_USER_POOL_ID not set")
	}

	userPoolClientID := os.Getenv("COGNITO_USER_POOL_CLIENT_ID")
	if userPoolClientID == "" {
		log.Println("WARNING: COGNITO_USER_POOL_CLIENT_ID not set")
	}

	cognitoRegion := os.Getenv("COGNITO_REGION")
	if cognitoRegion == "" {
		cognitoRegion = os.Getenv("AWS_REGION")
		if cognitoRegion == "" {
			cognitoRegion = "us-east-1" // Default to us-east-1
			log.Printf("WARNING: COGNITO_REGION and AWS_REGION not set, defaulting to %s", cognitoRegion)
		}
	}

	return &Config{
		ServerAddress:     serverAddress,
		CorsOrigins:       corsOrigins,
		Environment:       environment,
		UseDynamoDB:       useDynamoDB,
		DynamoDBTableName: dynamoDBTableName,
		UserPoolID:        userPoolID,
		UserPoolClientID:  userPoolClientID,
		CognitoRegion:     cognitoRegion,
	}
}
