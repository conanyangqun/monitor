package channels

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/wcy-dt/ponghub/internal/common/params"
	"github.com/wcy-dt/ponghub/internal/types/structures/configure"
)

// WebhookNotifier implements generic webhook notifications
type WebhookNotifier struct {
	config *configure.WebhookConfig
}

// NewWebhookNotifier creates a new generic webhook notifier
func NewWebhookNotifier(config *configure.WebhookConfig) *WebhookNotifier {
	return &WebhookNotifier{config: config}
}

// Send sends a generic webhook notification with enhanced configuration support
func (w *WebhookNotifier) Send(title, message string) error {
	// Create parameter resolver for processing Special Parameters
	resolver := params.NewParameterResolver()

	url := w.config.URL
	if url == "" {
		url = os.Getenv("WEBHOOK_URL")
	}

	if url == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	// Resolve Special Parameters in the URL
	url = resolver.ResolveParameters(url)

	// Prepare the payload
	payload, contentType, err := w.buildPayload(title, message)
	if err != nil {
		return fmt.Errorf("failed to build webhook payload: %v", err)
	}

	// Determine method
	method := "POST"
	if w.config.Method != "" {
		method = strings.ToUpper(w.config.Method)
	}

	// Prepare headers and resolve Special Parameters in header values
	headers := make(map[string]string)
	for key, value := range w.config.Headers {
		resolvedValue := resolver.ResolveParameters(value)
		headers[key] = resolvedValue
	}

	// Set authentication if configured
	if w.config.AuthType != "" {
		w.setAuthentication(headers, resolver)
	}

	// Execute request with retry logic
	return w.sendWithRetry(url, method, payload, contentType, headers)
}

// buildPayload constructs the webhook payload based on configuration
func (w *WebhookNotifier) buildPayload(title, message string) (interface{}, string, error) {
	// Create parameter resolver for processing Special Parameters
	resolver := params.NewParameterResolver()

	// Resolve Special Parameters in title and message
	resolvedTitle := resolver.ResolveParameters(title)
	resolvedMessage := resolver.ResolveParameters(message)

	data := map[string]interface{}{
		"title":     resolvedTitle,
		"message":   resolvedMessage,
		"Title":     resolvedTitle,   // Add uppercase version for template compatibility
		"Message":   resolvedMessage, // Add uppercase version for template compatibility
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "ponghub",
	}

	// Check for custom payload configuration first
	if w.config.CustomPayload != nil {
		return w.buildCustomPayload(data, resolver)
	}

	// Default JSON payload
	return data, "application/json", nil
}

// buildCustomPayload builds payload using custom payload configuration
func (w *WebhookNotifier) buildCustomPayload(data map[string]interface{}, resolver *params.ParameterResolver) (interface{}, string, error) {
	customPayload := w.config.CustomPayload

	// Create enhanced data with custom fields and field mappings
	enhancedData := make(map[string]interface{})

	// Copy original data
	for k, v := range data {
		enhancedData[k] = v
	}

	// Add custom fields if configured, resolving Special Parameters
	if customPayload.Fields != nil {
		for key, value := range customPayload.Fields {
			resolvedValue := resolver.ResolveParameters(value)
			enhancedData[key] = resolvedValue
		}
	}

	// Handle field name mapping
	if customPayload.TitleField != "" && customPayload.IncludeTitle {
		enhancedData[customPayload.TitleField] = data["title"]
	}
	if customPayload.MessageField != "" && customPayload.IncludeMessage {
		enhancedData[customPayload.MessageField] = data["message"]
	}

	// Use custom template if provided
	if customPayload.Template != "" {
		// Always use template parsing for custom payload templates
		return w.buildTemplatePayloadWithData(customPayload.Template, enhancedData, customPayload.ContentType, resolver)
	}

	// Default behavior - return enhanced data
	contentType := "application/json"
	if customPayload.ContentType != "" {
		contentType = customPayload.ContentType
	}

	return enhancedData, contentType, nil
}

// buildTemplatePayloadWithData builds payload using a template with provided data and content type
func (w *WebhookNotifier) buildTemplatePayloadWithData(templateStr string, data map[string]interface{}, contentType string, resolver *params.ParameterResolver) (interface{}, string, error) {
	// Create a custom resolver that preserves Go template syntax
	// We need to be careful not to process Go template variables like {{.Title}}
	resolvedTemplate := w.resolveSpecialParametersOnly(templateStr, resolver)

	// For JSON templates, we need to properly escape string values
	// Create a new data map with JSON-safe string values
	templateData := make(map[string]interface{})
	for k, v := range data {
		templateData[k] = v
	}

	// Create template
	tmpl := template.New("webhook")
	tmpl, err := tmpl.Parse(resolvedTemplate)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return nil, "", fmt.Errorf("failed to execute template: %w", err)
	}

	templateResult := buf.String()

	// Try to parse as JSON first
	var jsonData interface{}
	if err := json.Unmarshal([]byte(templateResult), &jsonData); err != nil {
		// If JSON parsing fails, it's likely because string values weren't properly escaped
		// Let's try a different approach - build the JSON structure directly
		if strings.Contains(resolvedTemplate, "{{.Title}}") && strings.Contains(resolvedTemplate, "{{.Message}}") {
			// This looks like a JSON template, let's build it properly
			jsonStruct := make(map[string]interface{})

			// Extract title and message values
			if title, ok := data["Title"].(string); ok {
				jsonStruct["alert"] = title
			}
			if message, ok := data["Message"].(string); ok {
				jsonStruct["details"] = message
			}

			// Add other fields from the template if they exist
			for key, value := range data {
				if key != "Title" && key != "Message" && key != "title" && key != "message" {
					jsonStruct[key] = value
				}
			}

			resultContentType := "application/json"
			if contentType != "" {
				resultContentType = contentType
			}
			return jsonStruct, resultContentType, nil
		}

		// If not a recognizable JSON template, return as string
		resultContentType := "text/plain"
		if contentType != "" {
			resultContentType = contentType
		}
		return templateResult, resultContentType, nil
	}

	// If template produced valid JSON, merge with any additional fields from data
	if jsonMap, ok := jsonData.(map[string]interface{}); ok {
		// Add any fields from data that aren't in the template result
		for key, value := range data {
			// Skip the standard template fields but include custom fields
			if key != "title" && key != "message" && key != "Title" && key != "Message" && key != "timestamp" && key != "service" {
				if _, exists := jsonMap[key]; !exists {
					jsonMap[key] = value
				}
			}
		}
		resultContentType := "application/json"
		if contentType != "" {
			resultContentType = contentType
		}
		return jsonMap, resultContentType, nil
	}

	resultContentType := "application/json"
	if contentType != "" {
		resultContentType = contentType
	}
	return jsonData, resultContentType, nil
}

// resolveSpecialParametersOnly resolves only Special Parameters while preserving Go template syntax
func (w *WebhookNotifier) resolveSpecialParametersOnly(templateStr string, resolver *params.ParameterResolver) string {
	// Use regex to find Special Parameters but exclude Go template variables
	// Go template variables start with {{. while Special Parameters don't
	re := regexp.MustCompile(`\{\{([^.}][^}]*)}}`)

	result := templateStr
	matches := re.FindAllStringSubmatchIndex(templateStr, -1)

	// Process matches from right to left to preserve indices
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		fullMatchStart, fullMatchEnd := match[0], match[1]
		paramStart, paramEnd := match[2], match[3]

		// Extract the parameter content
		param := strings.TrimSpace(templateStr[paramStart:paramEnd])

		// Resolve the Special Parameter
		resolvedValue := resolver.ResolveParameters("{{" + param + "}}")

		// Replace the match in the result
		result = result[:fullMatchStart] + resolvedValue + result[fullMatchEnd:]
	}

	return result
}

// setAuthentication sets authentication headers based on configuration
func (w *WebhookNotifier) setAuthentication(headers map[string]string, resolver *params.ParameterResolver) {
	switch strings.ToLower(w.config.AuthType) {
	case "bearer":
		if w.config.AuthToken != "" {
			resolvedToken := resolver.ResolveParameters(w.config.AuthToken)
			headers["Authorization"] = "Bearer " + resolvedToken
		}
	case "basic":
		if w.config.AuthUsername != "" && w.config.AuthPassword != "" {
			resolvedUsername := resolver.ResolveParameters(w.config.AuthUsername)
			resolvedPassword := resolver.ResolveParameters(w.config.AuthPassword)
			auth := fmt.Sprintf("%s:%s", resolvedUsername, resolvedPassword)
			headers["Authorization"] = "Basic " + w.base64Encode(auth)
		}
	case "apikey":
		if w.config.AuthToken != "" {
			resolvedToken := resolver.ResolveParameters(w.config.AuthToken)
			if w.config.AuthHeader != "" {
				resolvedHeader := resolver.ResolveParameters(w.config.AuthHeader)
				headers[resolvedHeader] = resolvedToken
			} else {
				headers["X-API-Key"] = resolvedToken
			}
		}
	}
}

// base64Encode encodes string to base64
func (w *WebhookNotifier) base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// sendWithRetry sends the webhook with retry logic
func (w *WebhookNotifier) sendWithRetry(url, method string, payload interface{}, contentType string, headers map[string]string) error {
	maxRetries := 0
	if w.config.Retries > 0 {
		maxRetries = w.config.Retries
	}

	timeout := 30
	if w.config.Timeout > 0 {
		timeout = w.config.Timeout
	}

	// Handle different payload types
	var body io.Reader
	if payload != nil {
		switch v := payload.(type) {
		case string:
			body = strings.NewReader(v)
		default:
			jsonData, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to marshal payload: %w", err)
			}
			body = bytes.NewReader(jsonData)
		}
	}

	return sendHTTPRequestWithCustomBody(url, method, body, contentType, headers, maxRetries, timeout, w.config.SkipTLSVerify)
}

// WebhookError represents a webhook-specific error
type WebhookError struct {
	StatusCode int
	Body       string
	Retryable  bool
	Message    string
}
