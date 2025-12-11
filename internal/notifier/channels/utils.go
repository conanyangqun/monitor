package channels

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPError represents an HTTP error with additional context
type HTTPError struct {
	StatusCode int
	Body       string
	URL        string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP request to %s failed with status %d: %s", e.URL, e.StatusCode, e.Body)
}

// IsRetryable checks if the HTTP error is retryable
func (e *HTTPError) IsRetryable() bool {
	// 5xx server errors and some 4xx client errors are retryable
	retryableCodes := []int{408, 429, 500, 502, 503, 504}
	for _, code := range retryableCodes {
		if e.StatusCode == code {
			return true
		}
	}
	return false
}

// HTTPClient creates an HTTP client with optional TLS configuration
func createHTTPClient(timeout int, skipTLSVerify bool) *http.Client {
	if timeout <= 0 {
		timeout = 30 // Default 30 seconds
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerify,
		},
	}

	return &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}
}

// SendHTTPRequest sends an HTTP request with retry logic
func sendHTTPRequest(url string, method string, payload interface{}, headers map[string]string, maxRetries, timeout int, skipTLSVerify bool) error {
	client := createHTTPClient(timeout, skipTLSVerify)

	var bodyReader io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry (exponential backoff)
			waitTime := time.Duration(attempt) * time.Second
			if waitTime > 10*time.Second {
				waitTime = 10 * time.Second
			}
			time.Sleep(waitTime)

			// Reset body reader for retry
			if payload != nil {
				jsonData, err := json.Marshal(payload)
				if err != nil {
					lastErr = fmt.Errorf("failed to marshal payload during retry: %w", err)
					continue
				}
				bodyReader = bytes.NewBuffer(jsonData)
			}
		}

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		// Set default content type if payload exists
		if payload != nil && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		// Add custom headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Read response body
		body, _ := io.ReadAll(resp.Body)
		if err := resp.Body.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close response body: %w", err)
		}

		// Check if request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		// Handle specific status codes
		switch resp.StatusCode {
		case 429: // Rate limited - always retry
			lastErr = fmt.Errorf("rate limited (429), response: %s", string(body))
		case 500, 502, 503, 504: // Server errors - retry
			lastErr = fmt.Errorf("server error (%d), response: %s", resp.StatusCode, string(body))
		default: // Client errors - don't retry
			return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
		}
	}

	return fmt.Errorf("request failed after %d retries, last error: %w", maxRetries+1, lastErr)
}

// SendHTTPRequestWithCustomBody sends an HTTP request with custom body content
func sendHTTPRequestWithCustomBody(url string, method string, body io.Reader, contentType string, headers map[string]string, maxRetries, timeout int, skipTLSVerify bool) error {
	client := createHTTPClient(timeout, skipTLSVerify)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			waitTime := time.Duration(attempt) * time.Second
			if waitTime > 10*time.Second {
				waitTime = 10 * time.Second
			}
			time.Sleep(waitTime)
		}

		req, err := http.NewRequest(method, url, body)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		// Set content type
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		// Add custom headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Read response body
		respBody, _ := io.ReadAll(resp.Body)
		if err := resp.Body.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close response body: %w", err)
		}

		// Check if request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		// Handle specific status codes
		switch resp.StatusCode {
		case 429: // Rate limited - always retry
			lastErr = fmt.Errorf("rate limited (429), response: %s", string(respBody))
		case 500, 502, 503, 504: // Server errors - retry
			lastErr = fmt.Errorf("server error (%d), response: %s", resp.StatusCode, string(respBody))
		default: // Client errors - don't retry
			return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
		}
	}

	return fmt.Errorf("request failed after %d retries, last error: %w", maxRetries+1, lastErr)
}
