package main

import (
	"log"

	"github.com/aws_e2e_test/usersvc/internal/config"
	"github.com/aws_e2e_test/usersvc/internal/usersvc"
)

func main() {
	// Load configuration from environment variables
	cfg := config.NewConfig()

	// Create and initialize the server
	server, err := usersvc.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start the server
	log.Printf("Starting server on %s", cfg.ServerAddress)
	if err := server.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
