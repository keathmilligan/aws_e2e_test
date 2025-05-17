package model

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a message in the system
type Message struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// NewMessage creates a new message with the given text
func NewMessage(text string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		Text:      text,
		Timestamp: time.Now(),
	}
}
