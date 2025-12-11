package channels

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wcy-dt/ponghub/internal/types/structures/configure"
)

func TestNewEmailNotifier(t *testing.T) {
	config := &configure.EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		From:     "test@example.com",
		To:       []string{"recipient@example.com"},
	}

	notifier := NewEmailNotifier(config)

	if notifier == nil {
		t.Fatal("NewEmailNotifier returned nil")
	}

	if notifier.config != config {
		t.Error("NewEmailNotifier did not set config correctly")
	}
}

func TestEmailNotifier_BuildEmailBody(t *testing.T) {
	config := &configure.EmailConfig{
		From:    "sender@example.com",
		To:      []string{"recipient1@example.com", "recipient2@example.com"},
		ReplyTo: "reply@example.com",
	}

	notifier := NewEmailNotifier(config)
	body := notifier.buildEmailBody("Test Subject", "Test Message")

	// Check that essential headers are present
	if !strings.Contains(body, "From: sender@example.com") {
		t.Error("Email body missing From header")
	}

	if !strings.Contains(body, "To: recipient1@example.com, recipient2@example.com") {
		t.Error("Email body missing or incorrect To header")
	}

	if !strings.Contains(body, "Subject: Test Subject") {
		t.Error("Email body missing Subject header")
	}

	if !strings.Contains(body, "Reply-To: reply@example.com") {
		t.Error("Email body missing Reply-To header")
	}

	if !strings.Contains(body, "MIME-Version: 1.0") {
		t.Error("Email body missing MIME-Version header")
	}

	if !strings.Contains(body, "Content-Type: text/plain; charset=UTF-8") {
		t.Error("Email body missing Content-Type header")
	}

	if !strings.Contains(body, "Test Message") {
		t.Error("Email body missing message content")
	}

	// Check that Date header is present and properly formatted
	if !strings.Contains(body, "Date: ") {
		t.Error("Email body missing Date header")
	}
}

func TestEmailNotifier_FormatRecipients(t *testing.T) {
	tests := []struct {
		name       string
		recipients []string
		expected   string
	}{
		{
			name:       "Single recipient",
			recipients: []string{"test@example.com"},
			expected:   "test@example.com",
		},
		{
			name:       "Multiple recipients",
			recipients: []string{"test1@example.com", "test2@example.com", "test3@example.com"},
			expected:   "test1@example.com, test2@example.com, test3@example.com",
		},
		{
			name:       "No recipients",
			recipients: []string{},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configure.EmailConfig{To: tt.recipients}
			notifier := NewEmailNotifier(config)
			result := notifier.formatRecipients()

			if result != tt.expected {
				t.Errorf("formatRecipients() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test for missing credentials
func TestEmailNotifier_Send_MissingCredentials(t *testing.T) {
	// Temporarily clear environment variables
	originalUsername := os.Getenv("SMTP_USERNAME")
	originalPassword := os.Getenv("SMTP_PASSWORD")

	_ = os.Unsetenv("SMTP_USERNAME")
	_ = os.Unsetenv("SMTP_PASSWORD")

	defer func() {
		if originalUsername != "" {
			_ = os.Setenv("SMTP_USERNAME", originalUsername)
		}
		if originalPassword != "" {
			_ = os.Setenv("SMTP_PASSWORD", originalPassword)
		}
	}()

	config := &configure.EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		From:     "test@example.com",
		To:       []string{"recipient@example.com"},
		UseTLS:   true,
	}

	notifier := NewEmailNotifier(config)
	err := notifier.Send("Test", "Test Message")

	if err == nil {
		t.Error("Expected error for missing credentials, but got nil")
		return
	}

	if !strings.Contains(err.Error(), "SMTP credentials not found") {
		t.Errorf("Expected credentials error, got: %v", err)
	}
}

// Integration test - requires environment variables to be set
func TestEmailNotifier_Send_Integration(t *testing.T) {
	// Skip integration test if environment variables are not set
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	testRecipient := os.Getenv("TEST_EMAIL_RECIPIENT")

	if username == "" || password == "" || testRecipient == "" {
		t.Skip("Skipping integration test: SMTP_USERNAME, SMTP_PASSWORD, and TEST_EMAIL_RECIPIENT environment variables must be set")
	}

	// Test with common email providers
	testConfigs := []struct {
		name   string
		config *configure.EmailConfig
	}{
		{
			name: "Gmail STARTTLS",
			config: &configure.EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    587,
				From:        username,
				To:          []string{testRecipient},
				UseStartTLS: true,
				SkipVerify:  false,
			},
		},
		{
			name: "Gmail TLS",
			config: &configure.EmailConfig{
				SMTPHost:   "smtp.gmail.com",
				SMTPPort:   465,
				From:       username,
				To:         []string{testRecipient},
				UseTLS:     true,
				SkipVerify: false,
			},
		},
		{
			name: "Outlook STARTTLS",
			config: &configure.EmailConfig{
				SMTPHost:    "smtp-mail.outlook.com",
				SMTPPort:    587,
				From:        username,
				To:          []string{testRecipient},
				UseStartTLS: true,
				SkipVerify:  false,
			},
		},
	}

	for _, tc := range testConfigs {
		t.Run(tc.name, func(t *testing.T) {
			notifier := NewEmailNotifier(tc.config)

			title := "PongHub Email Test - " + tc.name
			message := "This is a test email sent from PongHub email notifier.\n\n" +
				"Timestamp: " + time.Now().Format(time.RFC3339) + "\n" +
				"Test configuration: " + tc.name + "\n" +
				"If you receive this email, the email notification is working correctly.\n\n" +
				"Test Details:\n" +
				"- SMTP Host: " + tc.config.SMTPHost + "\n" +
				"- SMTP Port: " + strconv.Itoa(tc.config.SMTPPort) + "\n" +
				"- Use TLS: " + fmt.Sprintf("%t", tc.config.UseTLS) + "\n" +
				"- Use STARTTLS: " + fmt.Sprintf("%t", tc.config.UseStartTLS)

			err := notifier.Send(title, message)
			if err != nil {
				t.Errorf("Failed to send email with %s: %v", tc.name, err)
			} else {
				t.Logf("Successfully sent email with %s", tc.name)
			}
		})
	}
}

// Test with invalid SMTP server (should fail)
func TestEmailNotifier_Send_InvalidServer(t *testing.T) {
	// Set dummy credentials for this test
	_ = os.Setenv("SMTP_USERNAME", "test@example.com")
	_ = os.Setenv("SMTP_PASSWORD", "testpass")
	defer func() {
		_ = os.Unsetenv("SMTP_USERNAME")
		_ = os.Unsetenv("SMTP_PASSWORD")
	}()

	config := &configure.EmailConfig{
		SMTPHost:    "invalid.smtp.server.example",
		SMTPPort:    587,
		From:        "test@example.com",
		To:          []string{"recipient@example.com"},
		UseStartTLS: true,
	}

	notifier := NewEmailNotifier(config)
	err := notifier.Send("Test", "Test Message")

	if err == nil {
		t.Error("Expected error for invalid SMTP server, but got nil")
		return
	}

	// Should contain connection error
	if !strings.Contains(err.Error(), "connect") && !strings.Contains(err.Error(), "dial") {
		t.Logf("Got expected error: %v", err)
	}
}

// Benchmark test
func BenchmarkEmailNotifier_BuildEmailBody(b *testing.B) {
	config := &configure.EmailConfig{
		From: "sender@example.com",
		To:   []string{"recipient1@example.com", "recipient2@example.com"},
	}

	notifier := NewEmailNotifier(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		notifier.buildEmailBody("Benchmark Subject", "Benchmark message content for performance testing")
	}
}
