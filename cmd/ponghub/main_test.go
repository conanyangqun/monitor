package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/wcy-dt/ponghub/internal/checker"
	"github.com/wcy-dt/ponghub/internal/configure"
	"github.com/wcy-dt/ponghub/internal/logger"
	"github.com/wcy-dt/ponghub/internal/notifier"
	"github.com/wcy-dt/ponghub/internal/reporter"
	"github.com/wcy-dt/ponghub/internal/types/types/default_config"
)

// TestMain_append tests the main functionality when appending to an existing log file.
func TestMain_append(t *testing.T) {
	runMainFunctionality(true)
}

// TestMain_new tests the main functionality when creating a new log file.
func TestMain_new(t *testing.T) {
	runMainFunctionality(false)
}

// runMainFunctionality runs the main functionality for testing purposes.
// If copyExistingLog is true, it copies the existing log file to a temporary location.
func runMainFunctionality(copyExistingLog bool) {
	// load the default configuration
	cfg, err := configure.ReadConfigs(default_config.GetConfigPath())
	if err != nil {
		log.Fatalln("Error loading config at", default_config.GetConfigPath(), ":", err)
	}

	// copy log file to a temporary location for testing (only if copyExistingLog is true)
	if copyExistingLog {
		logPath := default_config.GetLogPath()
		if err := copyLogFile(logPath, tmpLogPath); err != nil {
			log.Fatalln("Error copying log file:", err)
		}
	}

	// check services based on the configuration
	checkResult := checker.CheckServices(cfg)

	// notify the result
	notifier.WriteNotifications(checkResult, cfg.CertNotifyDays)
	notifier.SendNotifications(checkResult, cfg.CertNotifyDays, cfg.Notifications)

	// get and write log results
	logResult, err := logger.GetLog(checkResult, cfg.MaxLogDays, tmpLogPath)
	if err != nil {
		log.Fatalln("Error outputting checkResult:", err)
	}
	if err := logger.WriteLog(logResult, tmpLogPath); err != nil {
		log.Fatalln("Error writing logs to", tmpLogPath, ":", err)
	} else {
		log.Println("Logs written to", tmpLogPath)
	}

	// generate the report based on the checkResult
	reportResult, err := reporter.GetReport(checkResult, tmpLogPath, cfg)
	if err != nil {
		log.Fatalln("Error generating report data:", err)
	}
	if err := reporter.WriteReport(reportResult, default_config.GetReportPath(), cfg.DisplayNum); err != nil {
		log.Fatalln("Error generating report:", err)
	} else {
		log.Println("Report generated at", default_config.GetReportPath())
	}

	// Remove the temporary log file after tests
	if err := os.Remove(tmpLogPath); err != nil {
		log.Println("Error removing temporary log file:", err)
	}
}

// copyLogFile copies the log file from srcPath to dstPath.
// If srcPath doesn't exist, it creates an empty JSON object file at dstPath.
func copyLogFile(srcPath, dstPath string) error {
	// remove dstPath if it exists
	if _, err := os.Stat(dstPath); err == nil {
		if err := os.Remove(dstPath); err != nil {
			log.Println("Error removing existing destination file:", err)
			return err
		}
	}

	// Check if source file exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		// Source file doesn't exist, create an empty JSON object file
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer func(dstFile *os.File) {
			if err := dstFile.Close(); err != nil {
				log.Println("Error closing destination file:", err)
			}
		}(dstFile)

		// Write empty JSON object
		_, err = dstFile.WriteString("{}")
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		if err := srcFile.Close(); err != nil {
			log.Println("Error closing source file:", err)
		}
	}(srcFile)

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func(dstFile *os.File) {
		if err := dstFile.Close(); err != nil {
			log.Println("Error closing destination file:", err)
		}
	}(dstFile)

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// tmpLogPath is a temporary log file path used for testing purposes.
const tmpLogPath = "data/ponghub_log_test.json"

func TestMain(m *testing.M) {
	// Change the working directory to the root of the project
	root, err := filepath.Abs("../..")
	if err != nil {
		panic(err)
	}
	if err := os.Chdir(root); err != nil {
		panic(err)
	}

	// Ensure data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}
