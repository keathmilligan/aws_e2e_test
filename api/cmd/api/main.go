package main

import (
	"log"
	"os"

	"github.com/awse2e/backend/internal/api"
	"github.com/awse2e/backend/internal/config"
)

func main() {
	// Get configuration from environment variables
	cfg := config.New()

	// Log storage configuration
	if cfg.UseDynamoDB {
		log.Printf("Storage configuration: DynamoDB (table: %s)", cfg.DynamoDBTableName)
	} else {
		log.Printf("Storage configuration: In-memory (local development mode)")
	}

	// Initialize the API server
	server := api.NewServer(cfg)

	// Start the server
	log.Printf("Starting server on %s", cfg.ServerAddress)
	if err := server.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
