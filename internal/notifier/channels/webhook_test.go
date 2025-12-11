package channels

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/wcy-dt/ponghub/internal/types/structures/configure"
)

// TestWebhookNotifier_BasicSend tests basic webhook functionality
//
//goland:noinspection DuplicatedCode
func TestWebhookNotifier_BasicSend(t *testing.T) {
	// Create a test server to receive webhook
	var receivedPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse JSON
		if err := json.Unmarshal(body, &receivedPayload); err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("OK"))
		if err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create webhook config
	config := &configure.WebhookConfig{
		URL:    server.URL,
		Method: "POST",
	}

	// Create notifier and send
	notifier := NewWebhookNotifier(config)
	err := notifier.Send("Test Alert", "This is a test message")

	if err != nil {
		t.Fatalf("Failed to send webhook: %v", err)
	}

	// Verify received data
	if receivedPayload["title"] != "Test Alert" {
		t.Errorf("Expected title 'Test Alert', got '%v'", receivedPayload["title"])
	}

	if receivedPayload["message"] != "This is a test message" {
		t.Errorf("Expected message 'This is a test message', got '%v'", receivedPayload["message"])
	}

	if receivedPayload["service"] != "ponghub" {
		t.Errorf("Expected service 'ponghub', got '%v'", receivedPayload["service"])
	}
}

// TestWebhookNotifier_CustomPayload tests the custom payload functionality
//
//goland:noinspection DuplicatedCode
func TestWebhookNotifier_CustomPayload(t *testing.T) {
	// Test server to capture requests
	var receivedPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse JSON
		if err := json.Unmarshal(body, &receivedPayload); err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create webhook config with custom payload
	config := &configure.WebhookConfig{
		URL:    server.URL,
		Method: "POST",
		CustomPayload: &configure.CustomPayloadConfig{
			Template:    `{"alert": "{{.Title}}", "details": "{{.Message}}", "env": "{{.environment}}"}`,
			ContentType: "application/json",
			Fields: map[string]string{
				"environment": "production",
			},
		},
	}

	notifier := NewWebhookNotifier(config)
	err := notifier.Send("Service Down", "Database connection failed")

	if err != nil {
		t.Fatalf("Failed to send webhook: %v", err)
	}

	// Verify the custom template was used correctly
	if receivedPayload["alert"] != "Service Down" {
		t.Errorf("Expected alert 'Service Down', got '%v'", receivedPayload["alert"])
	}

	if receivedPayload["details"] != "Database connection failed" {
		t.Errorf("Expected details 'Database connection failed', got '%v'", receivedPayload["details"])
	}

	if receivedPayload["env"] != "production" {
		t.Errorf("Expected env 'production', got '%v'", receivedPayload["env"])
	}
}

// TestWebhookNotifier_Authentication tests Bearer token authentication
func TestWebhookNotifier_Authentication(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &configure.WebhookConfig{
		URL:       server.URL,
		Method:    "POST",
		AuthType:  "bearer",
		AuthToken: "test-token-123",
	}

	notifier := NewWebhookNotifier(config)
	err := notifier.Send("Test", "Test message")

	if err != nil {
		t.Fatalf("Failed to send webhook: %v", err)
	}

	expectedAuth := "Bearer test-token-123"
	if receivedAuth != expectedAuth {
		t.Errorf("Expected Authorization '%s', got '%s'", expectedAuth, receivedAuth)
	}
}

// TestWebhookNotifier_ErrorHandling tests error handling and retries
func TestWebhookNotifier_ErrorHandling(t *testing.T) {
	// Test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Internal Server Error"))
		if err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	config := &configure.WebhookConfig{
		URL:     server.URL,
		Method:  "POST",
		Retries: 0, // No retries for this test
	}

	notifier := NewWebhookNotifier(config)
	err := notifier.Send("Test Alert", "Test message")

	if err == nil {
		t.Fatal("Expected error for 500 status code")
	}

	requestCount := 0
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config = &configure.WebhookConfig{
		URL:     server.URL,
		Method:  "POST",
		Retries: 3,
		Timeout: 5,
	}

	notifier = NewWebhookNotifier(config)
	err = notifier.Send("Test Alert", "Test message")

	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if requestCount != 3 {
		t.Errorf("Expected 3 requests (2 retries + 1 success), got %d", requestCount)
	}
}

// TestWebhookNotifier_ConcurrentRequests tests concurrent webhook sending
func TestWebhookNotifier_ConcurrentRequests(t *testing.T) {
	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &configure.WebhookConfig{
		URL:    server.URL,
		Method: "POST",
	}

	notifier := NewWebhookNotifier(config)

	const numWorkers = 5
	const requestsPerWorker = 2

	errChan := make(chan error, numWorkers*requestsPerWorker)

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for j := 0; j < requestsPerWorker; j++ {
				err := notifier.Send("Test", "Message")
				errChan <- err
			}
		}(i)
	}

	// Collect results
	var errors []error
	for i := 0; i < numWorkers*requestsPerWorker; i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		t.Fatalf("Got %d errors from concurrent requests: %v", len(errors), errors[0])
	}

	finalCount := atomic.LoadInt64(&requestCount)
	expectedCount := int64(numWorkers * requestsPerWorker)
	if finalCount != expectedCount {
		t.Errorf("Expected %d requests, got %d", expectedCount, finalCount)
	}
}

// TestWebhookNotifier_RealWorldScenario tests real-world webhook usage
func TestWebhookNotifier_RealWorldScenario(t *testing.T) {
	type AlertPayload struct {
		Alert   string `json:"alert"`
		Details string `json:"details"`
	}

	var receivedAlert AlertPayload
	var receivedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")

		body, _ := io.ReadAll(r.Body)

		// Parse the JSON body
		if err := json.Unmarshal(body, &receivedAlert); err != nil {
			t.Logf("JSON decode error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status": "success"}`))
		if err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	config := &configure.WebhookConfig{
		URL:    server.URL,
		Method: "POST",
		CustomPayload: &configure.CustomPayloadConfig{
			Template:    `{"alert": "{{.Title}}", "details": "{{.Message}}"}`,
			ContentType: "application/json",
		},
	}

	notifier := NewWebhookNotifier(config)

	title := "🔴 PongHub Service Status Alert"
	message := "Generated at: 2025-10-12 10:00:00\n\nService check failed"

	err := notifier.Send(title, message)
	if err != nil {
		t.Fatalf("Failed to send real-world webhook: %v", err)
	}

	// Verify that Content-Type was set correctly
	if receivedContentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", receivedContentType)
	}

	if receivedAlert.Alert != title {
		t.Errorf("Expected alert field '%s', got '%s'", title, receivedAlert.Alert)
	}

	if receivedAlert.Details != message {
		t.Errorf("Expected details field '%s', got '%s'", message, receivedAlert.Details)
	}
}

// TestWebhookNotifier_SpecialParametersInCustomPayload tests Special Parameters in custom payload
func TestWebhookNotifier_SpecialParametersInCustomPayload(t *testing.T) {
	var receivedPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &receivedPayload); err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &configure.WebhookConfig{
		URL:    server.URL,
		Method: "POST",
		CustomPayload: &configure.CustomPayloadConfig{
			Template: `{"alert": "{{.Title}}", "details": "{{.Message}}", "request_id": "{{uuid_short}}", "date": "{{%Y-%m-%d}}"}`,
			Fields: map[string]string{
				"environment": "test-{{rand(1,100)}}",
				"session_id":  "{{uuid}}",
			},
		},
	}

	notifier := NewWebhookNotifier(config)
	err := notifier.Send("Service Down", "Database error")

	if err != nil {
		t.Fatalf("Failed to send webhook with Special Parameters in custom payload: %v", err)
	}

	// Verify Special Parameters were resolved in template
	if _, exists := receivedPayload["request_id"]; !exists {
		t.Errorf("Expected 'request_id' field from template")
	}
	if _, exists := receivedPayload["date"]; !exists {
		t.Errorf("Expected 'date' field from template")
	}

	// Verify Special Parameters were resolved in custom fields
	environment, ok := receivedPayload["environment"].(string)
	if !ok {
		t.Errorf("Expected environment to be string, got %T: %v", receivedPayload["environment"], receivedPayload["environment"])
	} else if environment == "test-{{rand(1,100)}}" {
		t.Errorf("Special Parameters in environment field were not resolved")
	}

	sessionId, ok := receivedPayload["session_id"].(string)
	if !ok {
		t.Errorf("Expected session_id to be string, got %T: %v", receivedPayload["session_id"], receivedPayload["session_id"])
	} else if sessionId == "{{uuid}}" {
		t.Errorf("Special Parameters in session_id field were not resolved")
	}
}

// TestWebhookNotifier_SpecialParametersInAuth tests Special Parameters in authentication
func TestWebhookNotifier_SpecialParametersInAuth(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set a test environment variable
	t.Setenv("TEST_TOKEN", "secret-123")

	config := &configure.WebhookConfig{
		URL:       server.URL,
		Method:    "POST",
		AuthType:  "bearer",
		AuthToken: "{{env(TEST_TOKEN)}}-{{rand(1000,9999)}}",
	}

	notifier := NewWebhookNotifier(config)
	err := notifier.Send("Test", "Test message")

	if err != nil {
		t.Fatalf("Failed to send webhook with Special Parameters in auth: %v", err)
	}

	// Verify that the token contains the resolved environment variable
	if receivedAuth == "Bearer {{env(TEST_TOKEN)}}-{{rand(1000,9999)}}" {
		t.Errorf("Special Parameters in auth token were not resolved")
	}

	// Should start with "Bearer secret-123-" followed by a random number
	expectedPrefix := "Bearer secret-123-"
	if !contains(receivedAuth, expectedPrefix) {
		t.Errorf("Expected auth token to contain '%s', got '%s'", expectedPrefix, receivedAuth)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
