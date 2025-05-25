package usersvc

import (
	"log"
	"net/http"

	"github.com/aws_e2e_test/shared/auth"
	localauth "github.com/aws_e2e_test/usersvc/internal/auth"
	"github.com/aws_e2e_test/usersvc/internal/config"
	"github.com/aws_e2e_test/usersvc/internal/model"
	"github.com/aws_e2e_test/usersvc/internal/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// UserStore is an interface for user storage
type UserStore interface {
	GetByEmail(email string) (*model.User, error)
	GetAll() ([]*model.User, error)
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(email string) error
}

// Server represents the API server
type Server struct {
	router        *gin.Engine
	config        *config.Config
	userStore     UserStore
	cognitoClient *localauth.CognitoClient
	jwtValidator  *auth.JWTValidator
}

// NewServer creates a new API server
func NewServer(cfg *config.Config) (*Server, error) {
	var userStore UserStore
	var err error

	// Initialize the appropriate user store based on configuration
	if cfg.UseDynamoDB {
		userStore, err = store.NewDynamoDBUserStore(cfg.DynamoDBTableName)
		if err != nil {
			log.Printf("ERROR: Failed to create DynamoDB user store: %v", err)
			log.Printf("ERROR: Stack trace: %+v", err)
			log.Printf("CRITICAL: Falling back to in-memory user store (WARNING: not suitable for multiple instances)")
			userStore = store.NewUserStore()
		}
	} else {
		log.Println("STORAGE: Using in-memory user store (suitable for local development only)")
		log.Println("STORAGE: Set USE_DYNAMODB=true for production/multi-instance deployments")
		userStore = store.NewUserStore()
	}

	// Initialize Cognito client
	cognitoClient, err := localauth.NewCognitoClient(
		cfg.CognitoRegion,
		cfg.UserPoolID,
		cfg.UserPoolClientID,
	)
	if err != nil {
		log.Printf("ERROR: Failed to create Cognito client: %v", err)
		return nil, err
	}

	// Initialize JWT validator
	jwtValidator := auth.NewCognitoJWTValidator(cfg.CognitoRegion, cfg.UserPoolID)

	server := &Server{
		router:        gin.Default(),
		config:        cfg,
		userStore:     userStore,
		cognitoClient: cognitoClient,
		jwtValidator:  jwtValidator,
	}

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{cfg.CorsOrigins}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	server.router.Use(cors.New(corsConfig))

	// Register routes
	server.registerRoutes()

	return server, nil
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
		// Authentication endpoints
		api.POST("/auth/signup", s.signUp)
		api.POST("/auth/confirm", s.confirmSignUp)
		api.POST("/auth/resend-code", s.resendConfirmationCode)
		api.POST("/auth/login", s.login)
		api.POST("/auth/refresh", s.refreshToken)
		api.POST("/auth/forgot-password", s.forgotPassword)
		api.POST("/auth/confirm-forgot-password", s.confirmForgotPassword)

		// Protected user endpoints (require authentication)
		protected := api.Group("/users")
		protected.Use(auth.JWTAuthMiddleware(s.jwtValidator))
		{
			protected.GET("", s.getUsers)
			protected.GET("/:email", s.getUserByEmail)
			protected.POST("", s.createUser)
			protected.PUT("/:email", s.updateUser)
			protected.DELETE("/:email", s.deleteUser)
		}
	}
}

// signUp handles user registration
func (s *Server) signUp(c *gin.Context) {
	var request model.UserSignupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Sign up the user with Cognito
	err := s.cognitoClient.SignUp(
		request.Email,
		request.Password,
		request.FirstName,
		request.LastName,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign up user"})
		return
	}

	// Create the user in the database
	user := model.NewUser(request.Email, request.FirstName, request.LastName)
	err = s.userStore.Create(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user.ToResponse())
}

// confirmSignUp handles user registration confirmation
func (s *Server) confirmSignUp(c *gin.Context) {
	var request struct {
		Email            string `json:"email" binding:"required,email"`
		ConfirmationCode string `json:"confirmationCode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Confirm the user's registration with Cognito
	err := s.cognitoClient.ConfirmSignUp(request.Email, request.ConfirmationCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm sign up"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User confirmed successfully"})
}

// resendConfirmationCode resends the confirmation code to the user
func (s *Server) resendConfirmationCode(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Resend the confirmation code with Cognito
	err := s.cognitoClient.ResendConfirmationCode(request.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend confirmation code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Confirmation code resent successfully"})
}

// login handles user authentication
func (s *Server) login(c *gin.Context) {
	var request model.UserLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate the user with Cognito
	authResponse, err := s.cognitoClient.Login(request.Email, request.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// refreshToken refreshes the authentication tokens
func (s *Server) refreshToken(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Refresh the tokens with Cognito
	authResponse, err := s.cognitoClient.RefreshToken(request.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// forgotPassword initiates the forgot password flow
func (s *Server) forgotPassword(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Initiate the forgot password flow with Cognito
	err := s.cognitoClient.ForgotPassword(request.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate forgot password flow"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset code sent successfully"})
}

// confirmForgotPassword completes the forgot password flow
func (s *Server) confirmForgotPassword(c *gin.Context) {
	var request struct {
		Email            string `json:"email" binding:"required,email"`
		ConfirmationCode string `json:"confirmationCode" binding:"required"`
		NewPassword      string `json:"newPassword" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Confirm the forgot password with Cognito
	err := s.cognitoClient.ConfirmForgotPassword(
		request.Email,
		request.ConfirmationCode,
		request.NewPassword,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// getUsers returns all users
func (s *Server) getUsers(c *gin.Context) {
	users, err := s.userStore.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	// Convert users to response format
	responses := make([]*model.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
}

// getUserByEmail returns a user by email
func (s *Server) getUserByEmail(c *gin.Context) {
	email := c.Param("email")
	user, err := s.userStore.GetByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// createUser creates a new user
func (s *Server) createUser(c *gin.Context) {
	var request struct {
		Email     string `json:"email" binding:"required,email"`
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	existingUser, err := s.userStore.GetByEmail(request.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if user exists"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Create the user
	user := model.NewUser(request.Email, request.FirstName, request.LastName)
	err = s.userStore.Create(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user.ToResponse())
}

// updateUser updates an existing user
func (s *Server) updateUser(c *gin.Context) {
	email := c.Param("email")
	var request model.UserUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the existing user
	user, err := s.userStore.GetByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update the user fields
	if request.FirstName != "" {
		user.FirstName = request.FirstName
	}
	if request.LastName != "" {
		user.LastName = request.LastName
	}
	if request.Status != "" {
		user.Status = request.Status
	}
	user.UpdatedAt = model.NewUser("", "", "").UpdatedAt // Update the timestamp

	// Save the updated user
	err = s.userStore.Update(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// deleteUser deletes a user
func (s *Server) deleteUser(c *gin.Context) {
	email := c.Param("email")

	// Check if user exists
	user, err := s.userStore.GetByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete the user from Cognito
	err = s.cognitoClient.AdminDeleteUser(email)
	if err != nil {
		log.Printf("WARNING: Failed to delete user from Cognito: %v", err)
		// Continue with deleting from the database
	}

	// Delete the user from the database
	err = s.userStore.Delete(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
