package configure

type (
	// Service defines the configuration for a service, including its health and Endpoints ports
	Service struct {
		Name          string     `yaml:"name"`
		Endpoints     []Endpoint `yaml:"endpoints"`
		Timeout       int        `yaml:"timeout,omitempty"`
		MaxRetryTimes int        `yaml:"max_retry_times,omitempty"`
	}

	// Endpoint defines the configuration for a port
	Endpoint struct {
		URL                 string            `yaml:"url"`
		ParsedURL           string            `yaml:"-"`
		Method              string            `yaml:"method,omitempty"`
		Headers             map[string]string `yaml:"headers,omitempty"`
		ParsedHeaders       map[string]string `yaml:"-"`
		Body                string            `yaml:"body,omitempty"`
		ParsedBody          string            `yaml:"-"`
		StatusCode          int               `yaml:"status_code,omitempty"`
		ResponseRegex       string            `yaml:"response_regex,omitempty"`
		ParsedResponseRegex string            `yaml:"-"`
	}
)
