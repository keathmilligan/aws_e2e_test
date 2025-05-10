package e2e_tests

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	apiURL = flag.String("api-url", "", "The URL of the API service")
)

// Message represents a message from the API
type Message struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

func TestMain(m *testing.M) {
	flag.Parse()

	// If API URL is not provided, try to get it from environment variable
	if *apiURL == "" {
		*apiURL = os.Getenv("API_URL")
		if *apiURL == "" {
			fmt.Println("API URL must be provided via -api-url flag or API_URL environment variable")
			os.Exit(1)
		}
	}

	// Ensure API URL doesn't end with a slash
	if (*apiURL)[len(*apiURL)-1] == '/' {
		*apiURL = (*apiURL)[:len(*apiURL)-1]
	}

	fmt.Printf("Running end-to-end tests against API at: %s\n", *apiURL)

	// Run tests
	exitCode := m.Run()
	os.Exit(exitCode)
}

// TestHealthEndpoint tests the health endpoint of the API
func TestHealthEndpoint(t *testing.T) {
	t.Log("Testing health endpoint...")

	resp, err := http.Get(*apiURL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request to health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var healthResponse map[string]string
	if err := json.Unmarshal(body, &healthResponse); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	status, ok := healthResponse["status"]
	if !ok {
		t.Fatalf("Response does not contain 'status' field")
	}

	if status != "ok" {
		t.Fatalf("Expected status to be 'ok', got '%s'", status)
	}

	t.Log("Health endpoint test passed!")
}

// TestCreateAndGetMessages tests creating a new message and then retrieving it
func TestCreateAndGetMessages(t *testing.T) {
	t.Log("Testing message creation and retrieval...")

	// Create a new message
	messageText := fmt.Sprintf("Test message created at %s", time.Now().Format(time.RFC3339))
	message, err := createMessage(messageText)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	t.Logf("Created message with ID: %s", message.ID)

	// Add a short delay to ensure the message is stored in DynamoDB
	// This helps with eventual consistency in distributed systems
	t.Log("Waiting for message to be stored...")
	time.Sleep(2 * time.Second)

	// Get all messages
	messages, err := getMessages()
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	// Verify the created message is in the list
	found := false
	for _, m := range messages {
		t.Log("Checking message ID:", m.ID)
		t.Log("Checking message text:", m.Text)
		t.Log("Checking message timestamp:", m.Timestamp)
		if m.ID == message.ID {
			found = true
			if m.Text != messageText {
				t.Fatalf("Message text doesn't match. Expected '%s', got '%s'", messageText, m.Text)
			}
			break
		}
	}

	if !found {
		t.Fatalf("Created message with ID %s not found in the list of messages", message.ID)
	}

	t.Log("Message creation and retrieval test passed!")
}

// createMessage creates a new message with the given text
func createMessage(text string) (*Message, error) {
	requestBody, err := json.Marshal(map[string]string{
		"text": text,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := http.Post(*apiURL+"/messages", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &message, nil
}

// getMessages retrieves all messages from the API
func getMessages() ([]Message, error) {
	resp, err := http.Get(*apiURL + "/messages")
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var messages []Message
	if err := json.Unmarshal(body, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return messages, nil
}
