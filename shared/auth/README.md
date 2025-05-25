# Shared Authentication Library

This library provides shared JWT validation and middleware functionality for the AWS E2E Test project services.

## Features

- JWT token validation with JWKS support
- Gin middleware for JWT authentication
- Support for both generic JWT validation and AWS Cognito-specific validation
- Context helpers for extracting user information from JWT claims

## Usage

### JWT Validation

#### Generic JWT Validator

```go
import "github.com/aws_e2e_test/shared/auth"

// Create a generic JWT validator
validator := auth.NewJWTValidator(auth.JWTValidatorConfig{
    JWKSURL: "https://your-jwks-endpoint/.well-known/jwks.json",
    Issuer:  "https://your-issuer", // optional
})

// Validate a token
claims, err := validator.ValidateToken(tokenString)
if err != nil {
    // Handle validation error
}
```

#### AWS Cognito JWT Validator

```go
import "github.com/aws_e2e_test/shared/auth"

// Create a Cognito-specific JWT validator
validator := auth.NewCognitoJWTValidator("us-east-1", "your-user-pool-id")

// Validate a token
claims, err := validator.ValidateToken(tokenString)
if err != nil {
    // Handle validation error
}
```

### Gin Middleware

```go
import (
    "github.com/aws_e2e_test/shared/auth"
    "github.com/gin-gonic/gin"
)

// Create JWT validator
validator := auth.NewCognitoJWTValidator("us-east-1", "your-user-pool-id")

// Apply middleware to protected routes
router := gin.Default()
protected := router.Group("/api")
protected.Use(auth.JWTAuthMiddleware(validator))
{
    protected.GET("/users", getUsersHandler)
    protected.POST("/messages", createMessageHandler)
}
```

### Context Helpers

The middleware automatically extracts user information and stores it in the Gin context:

```go
func getUsersHandler(c *gin.Context) {
    // Get JWT claims
    claims, exists := auth.GetJWTClaimsFromContext(c)
    if !exists {
        // Handle missing claims
    }

    // Get user email
    email, exists := auth.GetUserEmailFromContext(c)
    if !exists {
        // Handle missing email
    }

    // Get user subject (unique ID)
    sub, exists := auth.GetUserSubFromContext(c)
    if !exists {
        // Handle missing subject
    }

    // Get access token
    token, exists := auth.GetAccessTokenFromContext(c)
    if !exists {
        // Handle missing token
    }
}
```

### Claim Extraction Helpers

```go
import "github.com/golang-jwt/jwt/v5"

// Extract user information from JWT claims
email, ok := auth.GetUserEmailFromClaims(claims)
username, ok := auth.GetUsernameFromClaims(claims)
sub, ok := auth.GetUserSubFromClaims(claims)
```

## Dependencies

- `github.com/gin-gonic/gin` - Web framework
- `github.com/golang-jwt/jwt/v5` - JWT library

## Integration

To use this library in your service:

1. Add the dependency to your `go.mod`:
```go
require (
    github.com/aws_e2e_test/shared/auth v0.0.0-00010101000000-000000000000
)

replace github.com/aws_e2e_test/shared/auth => ../shared/auth
```

2. Run `go mod tidy` to download dependencies

3. Import and use the library in your code as shown in the examples above