package configure

import (
	"log"
	"os"

	"github.com/wcy-dt/ponghub/internal/common/params"
	"github.com/wcy-dt/ponghub/internal/types/structures/configure"
	"github.com/wcy-dt/ponghub/internal/types/types/default_config"

	"gopkg.in/yaml.v3"
)

// ReadConfigs loads the configuration from a YAML file at the specified path
func ReadConfigs(path string) (*configure.Configure, error) {
	// Read the configuration file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Println("Error closing config file:", err)
		}
	}(f)

	// Decode the YAML configuration
	cfg := new(configure.Configure)
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(cfg); err != nil {
		log.Fatalln("Failed to decode YAML config:", err)
	}

	// Resolve dynamic parameters
	resolveConfigParameters(cfg)

	// Set default values for the configuration
	setDefaultConfigs(cfg)

	if len(cfg.Services) == 0 {
		log.Fatalln("No services defined in the configuration file")
	}
	return cfg, nil
}

// resolveConfigParameters resolves dynamic parameters in configuration
func resolveConfigParameters(cfg *configure.Configure) {
	resolver := params.NewParameterResolver()

	for i := range cfg.Services {
		for j := range cfg.Services[i].Endpoints {
			endpoint := &cfg.Services[i].Endpoints[j]

			// Resolve parameters
			endpoint.ParsedURL = resolver.ResolveParameters(endpoint.URL)
			endpoint.ParsedBody = resolver.ResolveParameters(endpoint.Body)
			endpoint.ParsedResponseRegex = resolver.ResolveParameters(endpoint.ResponseRegex)
			if endpoint.Headers != nil {
				endpoint.ParsedHeaders = make(map[string]string)
				for key, value := range endpoint.Headers {
					endpoint.ParsedHeaders[key] = resolver.ResolveParameters(value)
				}
			}
		}
	}
}

// setDefaultConfigs sets default values for the configuration fields
func setDefaultConfigs(cfg *configure.Configure) {
	default_config.SetDefaultTimeout(&cfg.Timeout)
	default_config.SetDefaultMaxRetryTimes(&cfg.MaxRetryTimes)
	default_config.SetDefaultMaxLogDays(&cfg.MaxLogDays)
	default_config.SetDefaultCertNotifyDays(&cfg.CertNotifyDays)
	default_config.SetDefaultDisplayNum(&cfg.DisplayNum)

	for i := range cfg.Services {
		default_config.SetDefaultTimeout(&cfg.Services[i].Timeout)
		default_config.SetDefaultMaxRetryTimes(&cfg.Services[i].MaxRetryTimes)
	}

	// Set default notification configuration
	setDefaultNotifications(cfg)
}

// setDefaultNotifications sets default values for notification configuration
func setDefaultNotifications(cfg *configure.Configure) {
	if cfg.Notifications == nil {
		// If no notifications configured, enable default (GitHub Actions exit 1)
		cfg.Notifications = &configure.NotificationConfig{
			Enabled: true,
			Methods: []string{"default"},
			Default: &configure.DefaultConfig{Enabled: true},
		}
		return
	}

	// Check if other notification methods are configured
	hasOtherMethods := false
	for _, method := range cfg.Notifications.Methods {
		if method == "email" || method == "webhook" {
			hasOtherMethods = true
			break
		}
	}

	// If default config is not set, set based on other methods
	if cfg.Notifications.Default == nil {
		if hasOtherMethods {
			// If other methods are configured, disable default by default
			cfg.Notifications.Default = &configure.DefaultConfig{Enabled: false}
		} else {
			// If no other methods, enable default
			cfg.Notifications.Default = &configure.DefaultConfig{Enabled: true}
		}
	}

	// If methods include default, ensure it's enabled
	for _, method := range cfg.Notifications.Methods {
		if method == "default" && cfg.Notifications.Default != nil {
			cfg.Notifications.Default.Enabled = true
		}
	}
}
