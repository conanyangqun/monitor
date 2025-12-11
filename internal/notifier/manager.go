package notifier

import (
	"fmt"
	"log"
	"strings"

	"github.com/wcy-dt/ponghub/internal/notifier/channels"
	"github.com/wcy-dt/ponghub/internal/types/structures/configure"
)

// NotificationManager manages multiple notification services
type NotificationManager struct {
	services []NotificationService
	config   *configure.NotificationConfig
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(config *configure.NotificationConfig) *NotificationManager {
	manager := &NotificationManager{
		config:   config,
		services: make([]NotificationService, 0),
	}

	// If no notification config is provided, use default method
	if config == nil {
		log.Println("No notification configuration found, using default GitHub Actions notification")
		defaultConfig := &configure.DefaultConfig{Enabled: true}
		manager.config = &configure.NotificationConfig{
			Enabled: true,
			Methods: []string{"default"},
			Default: defaultConfig,
		}
		manager.services = append(manager.services, channels.NewDefaultNotifier(defaultConfig))
		return manager
	}

	// If notifications are disabled, return empty manager
	if !config.Enabled {
		return &NotificationManager{}
	}

	// If no methods are specified but notifications are enabled, use default
	if len(config.Methods) == 0 {
		log.Println("Notifications enabled but no methods specified, using default GitHub Actions notification")
		if config.Default == nil {
			config.Default = &configure.DefaultConfig{Enabled: true}
		}
		config.Methods = []string{"default"}
		manager.services = append(manager.services, channels.NewDefaultNotifier(config.Default))
		return manager
	}

	// Initialize notification services based on configured methods
	for _, method := range config.Methods {
		switch strings.ToLower(method) {
		case "default":
			if config.Default == nil {
				config.Default = &configure.DefaultConfig{Enabled: true}
			}
			manager.services = append(manager.services, channels.NewDefaultNotifier(config.Default))
		case "email":
			if config.Email != nil {
				manager.services = append(manager.services, channels.NewEmailNotifier(config.Email))
			}
		case "webhook":
			if config.Webhook != nil {
				manager.services = append(manager.services, channels.NewWebhookNotifier(config.Webhook))
			}
		default:
			log.Printf("Unknown notification method: %s", method)
		}
	}

	return manager
}

// SendNotification sends notification through all configured services
func (nm *NotificationManager) SendNotification(title, message string) {
	if nm.config == nil || !nm.config.Enabled || len(nm.services) == 0 {
		log.Println("Notifications are disabled or no services configured")
		return
	}

	log.Printf("Sending notifications through %d service(s)", len(nm.services))

	var failedServices []string
	for i, service := range nm.services {
		if err := service.Send(title, message); err != nil {
			serviceName := nm.getServiceName(i)
			log.Printf("Failed to send notification via %s: %v", serviceName, err)
			failedServices = append(failedServices, serviceName)
		} else {
			serviceName := nm.getServiceName(i)
			log.Printf("Successfully sent notification via %s", serviceName)
		}
	}

	if len(failedServices) > 0 {
		log.Printf("Failed to send notifications via: %s", strings.Join(failedServices, ", "))
	}
}

// getServiceName returns the name of the service at the given index
func (nm *NotificationManager) getServiceName(index int) string {
	if index < len(nm.config.Methods) {
		return nm.config.Methods[index]
	}
	return fmt.Sprintf("service_%d", index)
}

// IsEnabled returns whether notifications are enabled
func (nm *NotificationManager) IsEnabled() bool {
	return nm.config != nil && nm.config.Enabled && len(nm.services) > 0
}
