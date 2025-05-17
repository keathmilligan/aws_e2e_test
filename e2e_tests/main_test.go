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
	t.Logf("Using API URL: %s", *apiURL)

	// Create a unique message text with timestamp and random suffix
	randomSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	messageText := fmt.Sprintf("Test message created at %s (ID: %s)",
		time.Now().Format(time.RFC3339), randomSuffix)

	t.Logf("Creating message with text: %s", messageText)
	message, err := createMessage(messageText)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	t.Logf("Created message with ID: %s", message.ID)

	// Add a longer delay to ensure the message is stored in DynamoDB
	// This helps with eventual consistency in distributed systems
	t.Log("Waiting for message to be stored...")
	time.Sleep(5 * time.Second) // Increased from 2 to 5 seconds

	// Get all messages
	t.Log("Retrieving all messages...")
	messages, err := getMessages()
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	t.Logf("Retrieved %d messages", len(messages))

	// Verify the created message is in the list
	found := false
	for i, m := range messages {
		t.Logf("Message %d - ID: %s, Text: %s, Timestamp: %s",
			i, m.ID, m.Text, m.Timestamp.Format(time.RFC3339))

		if m.ID == message.ID {
			t.Logf("Found matching message with ID: %s", m.ID)
			found = true
			if m.Text != messageText {
				t.Fatalf("Message text doesn't match. Expected '%s', got '%s'", messageText, m.Text)
			}
			break
		}
	}

	if !found {
		// Try one more time with a longer delay
		t.Log("Message not found, waiting longer and trying again...")
		time.Sleep(5 * time.Second)

		messages, err = getMessages()
		if err != nil {
			t.Fatalf("Failed to get messages on second attempt: %v", err)
		}

		t.Logf("Retrieved %d messages on second attempt", len(messages))

		for i, m := range messages {
			t.Logf("Second attempt - Message %d - ID: %s, Text: %s",
				i, m.ID, m.Text)

			if m.ID == message.ID {
				t.Logf("Found matching message with ID: %s on second attempt", m.ID)
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("Created message with ID %s not found in the list of messages after multiple attempts", message.ID)
		}
	}

	t.Log("Message creation and retrieval test passed!")
}

// createMessage creates a new message with the given text
func createMessage(text string) (*Message, error) {
	fmt.Printf("Creating message with text: %s\n", text)

	requestBody, err := json.Marshal(map[string]string{
		"text": text,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	fmt.Printf("POST request to: %s/messages\n", *apiURL)
	fmt.Printf("Request body: %s\n", string(requestBody))

	// Create a custom HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", *apiURL+"/messages", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("expected status code %d, got %d. Response: %s",
			http.StatusCreated, resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("Response body: %s\n", string(body))

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	fmt.Printf("Created message with ID: %s\n", message.ID)

	return &message, nil
}

// getMessages retrieves all messages from the API
func getMessages() ([]Message, error) {
	// Add a cache-busting query parameter to prevent caching
	cacheBuster := fmt.Sprintf("nocache=%d", time.Now().UnixNano())
	url := fmt.Sprintf("%s/messages?%s", *apiURL, cacheBuster)

	fmt.Printf("Fetching messages from: %s\n", url)

	// Create a custom HTTP client with no caching
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to prevent caching
	req.Header.Add("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Expires", "0")

	resp, err := client.Do(req)
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

	fmt.Printf("Response body: %s\n", string(body))

	var messages []Message
	if err := json.Unmarshal(body, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	fmt.Printf("Unmarshalled %d messages\n", len(messages))

	return messages, nil
}
