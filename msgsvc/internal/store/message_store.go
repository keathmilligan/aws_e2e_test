package store

import (
	"sync"

	"github.com/aws_e2e_test/msgsvc/internal/model"
)

// MessageStore is an in-memory store for messages
type MessageStore struct {
	messages []*model.Message
	mutex    sync.RWMutex
}

// NewMessageStore creates a new message store
func NewMessageStore() *MessageStore {
	return &MessageStore{
		messages: make([]*model.Message, 0),
	}
}

// GetAll returns all messages
func (s *MessageStore) GetAll() ([]*model.Message, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy of the messages to avoid race conditions
	result := make([]*model.Message, len(s.messages))
	copy(result, s.messages)
	return result, nil
}

// Add adds a new message to the store
func (s *MessageStore) Add(message *model.Message) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.messages = append(s.messages, message)
	return nil
}
