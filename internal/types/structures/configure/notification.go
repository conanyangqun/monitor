package configure

type (
	// NotificationConfig defines the configuration for all notification channels
	NotificationConfig struct {
		Enabled bool           `yaml:"enabled,omitempty"`
		Methods []string       `yaml:"methods,omitempty"`
		Default *DefaultConfig `yaml:"default,omitempty"`
		Email   *EmailConfig   `yaml:"email,omitempty"`
		Webhook *WebhookConfig `yaml:"webhook,omitempty"`
	}

	// EmailConfig defines SMTP email notification settings
	EmailConfig struct {
		SMTPHost    string   `yaml:"smtp_host"`
		SMTPPort    int      `yaml:"smtp_port"`
		From        string   `yaml:"from"`
		To          []string `yaml:"to"`
		ReplyTo     string   `yaml:"reply_to,omitempty"`
		UseTLS      bool     `yaml:"use_tls,omitempty"`
		UseStartTLS bool     `yaml:"use_starttls,omitempty"`
		SkipVerify  bool     `yaml:"skip_verify,omitempty"`
	}

	// CustomPayloadConfig defines custom payload configuration for webhooks
	CustomPayloadConfig struct {
		Template       string            `yaml:"template,omitempty"`
		ContentType    string            `yaml:"content_type,omitempty"`
		Fields         map[string]string `yaml:"fields,omitempty"`
		IncludeTitle   bool              `yaml:"include_title,omitempty"`
		IncludeMessage bool              `yaml:"include_message,omitempty"`
		TitleField     string            `yaml:"title_field,omitempty"`
		MessageField   string            `yaml:"message_field,omitempty"`
	}

	// WebhookConfig defines generic webhook notification settings
	WebhookConfig struct {
		URL           string               `yaml:"url,omitempty"`
		Method        string               `yaml:"method,omitempty"`
		Headers       map[string]string    `yaml:"headers,omitempty"`
		ContentType   string               `yaml:"content_type,omitempty"`
		CustomPayload *CustomPayloadConfig `yaml:"custom_payload,omitempty"`
		AuthType      string               `yaml:"auth_type,omitempty"`
		AuthToken     string               `yaml:"auth_token,omitempty"`
		AuthUsername  string               `yaml:"auth_username,omitempty"`
		AuthPassword  string               `yaml:"auth_password,omitempty"`
		AuthHeader    string               `yaml:"auth_header,omitempty"`
		Retries       int                  `yaml:"retries,omitempty"`
		Timeout       int                  `yaml:"timeout,omitempty"`
		SkipTLSVerify bool                 `yaml:"skip_tls_verify,omitempty"`
	}
)
