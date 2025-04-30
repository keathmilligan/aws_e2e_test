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

	// Initialize the API server
	server := api.NewServer(cfg)

	// Start the server
	log.Printf("Starting server on %s", cfg.ServerAddress)
	if err := server.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
