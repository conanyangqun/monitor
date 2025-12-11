package channels

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/wcy-dt/ponghub/internal/types/structures/configure"
)

// EmailNotifier implements email notifications
type EmailNotifier struct {
	config *configure.EmailConfig
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(config *configure.EmailConfig) *EmailNotifier {
	return &EmailNotifier{config: config}
}

// Send sends an email notification with secure SMTP connection
func (e *EmailNotifier) Send(title, message string) error {
	// Get SMTP credentials from environment variables
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	if username == "" || password == "" {
		return fmt.Errorf("SMTP credentials not found in environment variables")
	}

	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)

	// Use secure connection based on configuration
	if e.config.UseTLS {
		// Direct TLS connection (typically port 465)
		return e.sendWithTLS(addr, username, password, title, message)
	} else if e.config.UseStartTLS {
		// STARTTLS connection (typically port 587)
		return e.sendWithStartTLS(addr, username, password, title, message)
	} else {
		// Plain connection - warn about security risk
		fmt.Printf("WARNING: Using plain SMTP connection without TLS. This is insecure and credentials will be sent in plain text. Consider enabling use_tls or use_starttls in your configuration.\n")
		return e.sendPlain(addr, username, password, title, message)
	}
}

// sendWithTLS sends email using direct TLS connection
func (e *EmailNotifier) sendWithTLS(addr, username, password, title, message string) error {
	tlsConfig := &tls.Config{
		ServerName:         e.config.SMTPHost,
		InsecureSkipVerify: e.config.SkipVerify,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to establish TLS connection: %w", err)
	}
	defer func(conn *tls.Conn) {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing TLS connection: %v", err)
		}
	}(conn)

	client, err := smtp.NewClient(conn, e.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer func(client *smtp.Client) {
		if err := client.Quit(); err != nil {
			log.Printf("Error quitting SMTP client: %v", err)
		}
	}(client)

	auth := smtp.PlainAuth("", username, password, e.config.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	return e.sendMessage(client, title, message)
}

// sendWithStartTLS sends email using STARTTLS
func (e *EmailNotifier) sendWithStartTLS(addr, username, password, title, message string) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer func(client *smtp.Client) {
		if err := client.Quit(); err != nil {
			log.Printf("Error quitting SMTP client: %v", err)
		}
	}(client)

	tlsConfig := &tls.Config{
		ServerName:         e.config.SMTPHost,
		InsecureSkipVerify: e.config.SkipVerify,
	}

	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	auth := smtp.PlainAuth("", username, password, e.config.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	return e.sendMessage(client, title, message)
}

// sendPlain sends email using plain connection (not recommended)
func (e *EmailNotifier) sendPlain(addr, username, password, title, message string) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer func(client *smtp.Client) {
		if err := client.Quit(); err != nil {
			log.Printf("Error quitting SMTP client: %v", err)
		}
	}(client)

	auth := smtp.PlainAuth("", username, password, e.config.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	return e.sendMessage(client, title, message)
}

// sendMessage sends the actual email message using the SMTP client
func (e *EmailNotifier) sendMessage(client *smtp.Client, title, message string) error {
	// Set sender
	if err := client.Mail(e.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range e.config.To {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send email body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	emailBody := e.buildEmailBody(title, message)
	if _, err := writer.Write([]byte(emailBody)); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close email writer: %w", err)
	}

	return nil
}

// buildEmailBody constructs the email body with proper headers
func (e *EmailNotifier) buildEmailBody(title, message string) string {
	headers := make(map[string]string)
	headers["From"] = e.config.From
	headers["To"] = e.formatRecipients()
	headers["Subject"] = title
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=UTF-8"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	// Add custom headers if configured
	if e.config.ReplyTo != "" {
		headers["Reply-To"] = e.config.ReplyTo
	}

	// Build header string
	headerStr := ""
	for key, value := range headers {
		headerStr += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	// Combine headers and body
	return headerStr + "\r\n" + message
}

// formatRecipients formats the recipient list for the To header
func (e *EmailNotifier) formatRecipients() string {
	if len(e.config.To) == 0 {
		return ""
	}

	result := ""
	for i, recipient := range e.config.To {
		if i > 0 {
			result += ", "
		}
		result += recipient
	}
	return result
}
