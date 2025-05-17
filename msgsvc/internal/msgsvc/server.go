package msgsvc

import (
	"log"
	"net/http"

	"github.com/aws_e2e_test/msgsvc/internal/config"
	"github.com/aws_e2e_test/msgsvc/internal/model"
	"github.com/aws_e2e_test/msgsvc/internal/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// MessageStore is an interface for message storage
type MessageStore interface {
	GetAll() ([]*model.Message, error)
	Add(message *model.Message) error
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
		messageStore, err = store.NewDynamoDBMessageStore(cfg.DynamoDBTableName)
		if err != nil {
			log.Printf("ERROR: Failed to create DynamoDB message store: %v", err)
			log.Printf("ERROR: Stack trace: %+v", err)
			log.Printf("CRITICAL: Falling back to in-memory message store (WARNING: not suitable for multiple instances)")
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
	// Add cache control headers to prevent caching
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	log.Printf("Handling GET /messages request")
	messages, err := s.messageStore.GetAll()
	if err != nil {
		log.Printf("Error getting messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	log.Printf("Returning %d messages", len(messages))
	for i, msg := range messages {
		log.Printf("Message %d: ID=%s, Text=%s", i, msg.ID, msg.Text)
	}

	c.JSON(http.StatusOK, messages)
}

// createMessage creates a new message
func (s *Server) createMessage(c *gin.Context) {
	log.Printf("Handling POST /messages request")

	var request struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Creating new message with text: %s", request.Text)
	message := model.NewMessage(request.Text)
	log.Printf("Generated message with ID: %s", message.ID)

	err := s.messageStore.Add(message)
	if err != nil {
		log.Printf("Error adding message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store message"})
		return
	}

	log.Printf("Successfully added message with ID: %s", message.ID)

	// Add cache control headers to prevent caching
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusCreated, message)
}
