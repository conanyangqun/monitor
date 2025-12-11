package logger

import (
	"log"

	"github.com/wcy-dt/ponghub/internal/common"
	"github.com/wcy-dt/ponghub/internal/types/structures/checker"
	"github.com/wcy-dt/ponghub/internal/types/structures/logger"
)

// GetLog writes check results to JSON file
func GetLog(currentCheckResult []checker.Service, maxLogDays int, logPath string) (logger.Logger, error) {
	// Load existing log data
	previousLog, err := common.ReadLogs(logPath)
	if err != nil {
		log.Printf("Error loading log data from %s: %v", logPath, err)
		return nil, err
	}

	// Use filtered data for further processing
	previousLog = common.FilterLogs(previousLog, currentCheckResult)

	// Merge new check results with existing log data
	currentLog := common.MergeLogs(previousLog, currentCheckResult, maxLogDays)

	return currentLog, nil
}

// WriteLog writes log data to file
func WriteLog(currentLog logger.Logger, logPath string) error {
	// Save updated log data back to file
	err := common.WriteLogs(currentLog, logPath)
	if err != nil {
		log.Printf("Error saving log data to %s: %v", logPath, err)
		return err
	}

	return nil
}
