package store

import (
	"github.com/aws_e2e_test/usersvc/internal/model"
)

// UserStore is an interface for user storage
type UserStore interface {
	// GetByEmail retrieves a user by email
	GetByEmail(email string) (*model.User, error)

	// GetAll retrieves all users
	GetAll() ([]*model.User, error)

	// Create creates a new user
	Create(user *model.User) error

	// Update updates an existing user
	Update(user *model.User) error

	// Delete deletes a user by email
	Delete(email string) error
}

// NewUserStore creates a new in-memory user store
func NewUserStore() UserStore {
	return &InMemoryUserStore{
		users: make(map[string]*model.User),
	}
}

// InMemoryUserStore is an in-memory implementation of UserStore
type InMemoryUserStore struct {
	users map[string]*model.User
}

// GetByEmail retrieves a user by email
func (s *InMemoryUserStore) GetByEmail(email string) (*model.User, error) {
	user, exists := s.users[email]
	if !exists {
		return nil, nil
	}
	return user, nil
}

// GetAll retrieves all users
func (s *InMemoryUserStore) GetAll() ([]*model.User, error) {
	users := make([]*model.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users, nil
}

// Create creates a new user
func (s *InMemoryUserStore) Create(user *model.User) error {
	s.users[user.Email] = user
	return nil
}

// Update updates an existing user
func (s *InMemoryUserStore) Update(user *model.User) error {
	s.users[user.Email] = user
	return nil
}

// Delete deletes a user by email
func (s *InMemoryUserStore) Delete(email string) error {
	delete(s.users, email)
	return nil
}
