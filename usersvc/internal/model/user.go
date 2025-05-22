package model

import (
	"time"
)

// User represents a user in the system
type User struct {
	Email     string    `json:"email" dynamodbav:"Email"`
	FirstName string    `json:"firstName" dynamodbav:"FirstName"`
	LastName  string    `json:"lastName" dynamodbav:"LastName"`
	Status    string    `json:"status" dynamodbav:"Status"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"CreatedAt"`
	UpdatedAt time.Time `json:"updatedAt" dynamodbav:"UpdatedAt"`
}

// UserStatus defines the possible status values for a user
type UserStatus string

const (
	// UserStatusActive indicates an active user
	UserStatusActive UserStatus = "ACTIVE"
	// UserStatusInactive indicates an inactive user
	UserStatusInactive UserStatus = "INACTIVE"
	// UserStatusPending indicates a pending user
	UserStatusPending UserStatus = "PENDING"
)

// NewUser creates a new user with the given details
func NewUser(email, firstName, lastName string) *User {
	now := time.Now()
	return &User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Status:    string(UserStatusActive),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UserSignupRequest represents the request to sign up a new user
type UserSignupRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
}

// UserLoginRequest represents the request to log in a user
type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserUpdateRequest represents the request to update a user
type UserUpdateRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Status    string `json:"status"`
}

// UserResponse represents the response for user operations
type UserResponse struct {
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ToResponse converts a User to a UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// AuthResponse represents the response for authentication operations
type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	IdToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}
