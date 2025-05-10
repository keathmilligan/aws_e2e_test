package api

import (
	"log"
	"net/http"

	"github.com/awse2e/backend/internal/config"
	"github.com/awse2e/backend/internal/model"
	"github.com/awse2e/backend/internal/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// MessageStore is an interface for message storage
type MessageStore interface {
	GetAll() []*model.Message
	Add(message *model.Message)
}

// Server represents the API server
type Server struct {
	router       *gin.Engine
	config       *config.Config
	messageStore MessageStore
}

// NewServer creates a new API server
func NewServer(cfg *config.Config) *Server {
	var messageStore MessageStore
	var err error

	// Initialize the appropriate message store based on configuration
	if cfg.UseDynamoDB {
		log.Println("STORAGE: Using DynamoDB message store for distributed persistence")
		messageStore, err = store.NewDynamoDBMessageStore(cfg.DynamoDBTableName)
		if err != nil {
			log.Printf("ERROR: Failed to create DynamoDB message store: %v", err)
			log.Println("STORAGE: Falling back to in-memory message store (WARNING: not suitable for multiple instances)")
			messageStore = store.NewMessageStore()
		}
	} else {
		log.Println("STORAGE: Using in-memory message store (suitable for local development only)")
		log.Println("STORAGE: Set USE_DYNAMODB=true for production/multi-instance deployments")
		messageStore = store.NewMessageStore()
	}

	server := &Server{
		router:       gin.Default(),
		config:       cfg,
		messageStore: messageStore,
	}

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{cfg.CorsOrigins}
	corsConfig.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type"}
	server.router.Use(cors.New(corsConfig))

	// Register routes
	server.registerRoutes()

	return server
}

// Run starts the server
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {
	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API endpoints
	api := s.router.Group("/")
	{
		// Get all messages
		api.GET("/messages", s.getMessages)

		// Create a new message
		api.POST("/messages", s.createMessage)
	}
}

// getMessages returns all messages
func (s *Server) getMessages(c *gin.Context) {
	messages := s.messageStore.GetAll()
	c.JSON(http.StatusOK, messages)
}

// createMessage creates a new message
func (s *Server) createMessage(c *gin.Context) {
	var request struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := model.NewMessage(request.Text)
	s.messageStore.Add(message)

	c.JSON(http.StatusCreated, message)
}
