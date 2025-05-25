package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware creates a middleware that validates JWT tokens
func JWTAuthMiddleware(jwtValidator *JWTValidator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get the Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			ctx.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must start with 'Bearer '"})
			ctx.Abort()
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
			ctx.Abort()
			return
		}

		// Validate the JWT token
		claims, err := jwtValidator.ValidateToken(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			ctx.Abort()
			return
		}

		// Store user information in the context for use in handlers
		ctx.Set("jwt_claims", claims)
		ctx.Set("access_token", token)

		// Extract common user info for convenience
		if email, ok := GetUserEmailFromClaims(claims); ok {
			ctx.Set("user_email", email)
		}
		if username, ok := GetUsernameFromClaims(claims); ok {
			ctx.Set("username", username)
		}
		if sub, ok := GetUserSubFromClaims(claims); ok {
			ctx.Set("user_sub", sub)
		}

		// Continue to the next handler
		ctx.Next()
	}
}

// GetJWTClaimsFromContext extracts JWT claims from the Gin context
func GetJWTClaimsFromContext(ctx *gin.Context) (map[string]interface{}, bool) {
	claims, exists := ctx.Get("jwt_claims")
	if !exists {
		return nil, false
	}

	claimsMap, ok := claims.(map[string]interface{})
	return claimsMap, ok
}

// GetUserEmailFromContext extracts the user email from the Gin context
func GetUserEmailFromContext(ctx *gin.Context) (string, bool) {
	email, exists := ctx.Get("user_email")
	if !exists {
		return "", false
	}

	emailStr, ok := email.(string)
	return emailStr, ok
}

// GetUserSubFromContext extracts the user subject (unique ID) from the Gin context
func GetUserSubFromContext(ctx *gin.Context) (string, bool) {
	sub, exists := ctx.Get("user_sub")
	if !exists {
		return "", false
	}

	subStr, ok := sub.(string)
	return subStr, ok
}

// GetAccessTokenFromContext extracts the access token from the Gin context
func GetAccessTokenFromContext(ctx *gin.Context) (string, bool) {
	token, exists := ctx.Get("access_token")
	if !exists {
		return "", false
	}

	tokenStr, ok := token.(string)
	return tokenStr, ok
}
