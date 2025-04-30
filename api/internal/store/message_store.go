package store

import (
	"sync"

	"github.com/awse2e/backend/internal/model"
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
func (s *MessageStore) GetAll() []*model.Message {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy of the messages to avoid race conditions
	result := make([]*model.Message, len(s.messages))
	copy(result, s.messages)
	return result
}

// Add adds a new message to the store
func (s *MessageStore) Add(message *model.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.messages = append(s.messages, message)
}
