package reporter

import (
	"log"
	"sort"

	"github.com/wcy-dt/ponghub/internal/types/structures/configure"
	"github.com/wcy-dt/ponghub/internal/types/structures/logger"
)

// convertToHistory converts logger history entries to reporter history format,
// sorts them by time and returns only the last displayNum entries.
func convertToHistory(logEntries []logger.HistoryEntry, displayNum int) History {
	var history History
	for _, entry := range logEntries {
		history = append(history, HistoryEntry{
			Time:         entry.Time,
			Status:       entry.Status,
			ResponseTime: entry.ResponseTime,
		})
	}

	// Sort the history by time, the most recent entries first
	sort.Slice(history, func(i, j int) bool {
		return history[i].Time < history[j].Time
	})

	// Get only the last `displayNum` entries
	if len(history) > displayNum {
		history = history[len(history)-displayNum:]
	}

	return history
}

// ParseLogResult converts logger.Logger data into a reporter.Reporter format preserving config order
func ParseLogResult(logResult logger.Logger, serviceNames []string, cfg *configure.Configure) Reporter {
	var report Reporter

	// Process services in the order they appear in config
	for _, serviceName := range serviceNames {
		serviceLog, exists := logResult[serviceName]
		if !exists {
			continue
		}

		if len(serviceLog.ServiceHistory) == 0 {
			log.Printf("No history data for service %s", serviceName)
			continue // Skip services with no history data
		}

		// Find the service config to get endpoint order
		var serviceConfig *configure.Service
		for _, svc := range cfg.Services {
			if svc.Name == serviceName {
				serviceConfig = &svc
				break
			}
		}

		// Convert logger.Endpoints to reporter.Endpoints preserving config order
		var endpoints Endpoints
		if serviceConfig != nil {
			// Process endpoints in config order
			for _, endpointConfig := range serviceConfig.Endpoints {
				url := endpointConfig.URL
				if endpointLog, exists := serviceLog.Endpoints[url]; exists {
					endpointHistory := convertToHistory(endpointLog, cfg.DisplayNum)
					endpoints = append(endpoints, Endpoint{
						URL:             url,
						EndpointHistory: endpointHistory,
					})
				}
			}
		} else {
			// Fallback: if no config found, use existing endpoints (shouldn't happen normally)
			for url, endpointLog := range serviceLog.Endpoints {
				endpointHistory := convertToHistory(endpointLog, cfg.DisplayNum)
				endpoints = append(endpoints, Endpoint{
					URL:             url,
					EndpointHistory: endpointHistory,
				})
			}
		}

		// convert logger.ServiceHistory to reporter.ServiceHistory
		serviceHistory := convertToHistory(serviceLog.ServiceHistory, cfg.DisplayNum)

		newService := Service{
			Name:           serviceName,
			ServiceHistory: serviceHistory,
			Endpoints:      endpoints,
		}
		report = append(report, newService)
	}
	return report
}
