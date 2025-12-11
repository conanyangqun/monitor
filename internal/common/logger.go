package common

import (
	"encoding/json"
	"os"

	"github.com/wcy-dt/ponghub/internal/types/structures/checker"
	"github.com/wcy-dt/ponghub/internal/types/structures/logger"
)

// ReadLogs loads log data from file or returns empty data
func ReadLogs(logPath string) (logger.Logger, error) {
	logResult := make(logger.Logger)

	logContent, err := os.ReadFile(logPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		return logResult, nil
	}

	if err := json.Unmarshal(logContent, &logResult); err != nil {
		return nil, err
	}
	return logResult, nil
}

// WriteLogs writes log data to file
func WriteLogs(logResult logger.Logger, logPath string) error {
	logContent, err := json.MarshalIndent(logResult, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(logPath, logContent, 0644)
}

// FilterLogs filters the previous log to include only services and endpoints present in the current check results
func FilterLogs(previousLog logger.Logger, currentCheckResult []checker.Service) logger.Logger {
	filteredPreviousLogs := make(logger.Logger)

	// Create maps for quick lookup of existing services and endpoints
	currentServices, currentEndpoints := getMapOfCurrentServicesAndEndpoints(currentCheckResult)

	// Filter the log data
	for serviceName, serviceLog := range previousLog {
		if currentServices[serviceName] {
			filteredPreviousLog := logger.Service{
				ServiceHistory: serviceLog.ServiceHistory,
				Endpoints:      make(logger.Endpoints),
			}

			// Filter endpoints for this service
			for endpointURL, endpointHistory := range serviceLog.Endpoints {
				if currentEndpoints[serviceName][endpointURL] {
					filteredPreviousLog.Endpoints[endpointURL] = endpointHistory
				}
			}

			// Only add the service if it has at least one endpoint
			if len(filteredPreviousLog.Endpoints) > 0 {
				filteredPreviousLogs[serviceName] = filteredPreviousLog
			}
		}
	}

	return filteredPreviousLogs
}

// getMapOfCurrentServicesAndEndpoints creates maps for quick lookup of existing services and endpoints
func getMapOfCurrentServicesAndEndpoints(currentCheckResult []checker.Service) (map[string]bool, map[string]map[string]bool) {
	// Create maps for quick lookup of existing services and endpoints
	currentServices := make(map[string]bool)
	currentEndpoints := make(map[string]map[string]bool) // serviceName -> endpoint URL -> exists

	for _, serviceResult := range currentCheckResult {
		currentServices[serviceResult.Name] = true
		currentEndpoints[serviceResult.Name] = make(map[string]bool)
		for _, endpoint := range serviceResult.Endpoints {
			currentEndpoints[serviceResult.Name][endpoint.URL] = true
		}
	}

	return currentServices, currentEndpoints
}

// MergeLogs merges previous log data with current check results and cleans up old entries
func MergeLogs(previousLog logger.Logger, currentCheckResult []checker.Service, maxLogDays int) logger.Logger {
	mergedLog := previousLog

	for _, serviceResult := range currentCheckResult {
		serviceName := serviceResult.Name

		serviceLog, exists := mergedLog[serviceName]
		if !exists {
			serviceLog = logger.Service{
				ServiceHistory: logger.History{},
				Endpoints:      make(logger.Endpoints),
			}
		}

		// Update service history
		newServiceHistoryEntry := logger.HistoryEntry{
			Time:   serviceResult.StartTime, // Use StartTime for the history entry
			Status: serviceResult.Status.String(),
		}
		serviceLog.ServiceHistory = serviceLog.ServiceHistory.AddEntry(newServiceHistoryEntry)
		serviceLog.ServiceHistory = serviceLog.ServiceHistory.CleanExpiredEntries(maxLogDays)

		// Update port statusList
		urlStatusMap, urlTimeMap, urlResponseTimeMap := processCheckResult(serviceResult)
		for url, statusList := range urlStatusMap {
			mergedStatus := calcMergedStatus(statusList)
			newEndpointHistoryEntry := logger.HistoryEntry{
				Time:         urlTimeMap[url],
				Status:       mergedStatus.String(),
				ResponseTime: int(urlResponseTimeMap[url].Milliseconds()),
			}

			tmp := serviceLog.Endpoints[url]
			tmp = tmp.AddEntry(newEndpointHistoryEntry)
			tmp = tmp.CleanExpiredEntries(maxLogDays)
			serviceLog.Endpoints[url] = tmp
		}

		mergedLog[serviceName] = serviceLog
	}

	return mergedLog
}
