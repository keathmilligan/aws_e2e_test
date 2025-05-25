package main

import (
	"log"
	"os"

	"github.com/aws_e2e_test/msgsvc/internal/config"
	"github.com/aws_e2e_test/msgsvc/internal/msgsvc"
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
	server, err := msgsvc.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
		os.Exit(1)
	}

	// Start the server
	log.Printf("Starting server on %s", cfg.ServerAddress)
	if err := server.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
