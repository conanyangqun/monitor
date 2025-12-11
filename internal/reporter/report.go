package reporter

import (
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/wcy-dt/ponghub/internal/common"
	"github.com/wcy-dt/ponghub/internal/types/structures/checker"
	"github.com/wcy-dt/ponghub/internal/types/structures/configure"
	"github.com/wcy-dt/ponghub/internal/types/structures/reporter"
	"github.com/wcy-dt/ponghub/internal/types/types/chk_result"
	"github.com/wcy-dt/ponghub/internal/types/types/default_config"
)

// GetReport generates a report based on the check results and log data
func GetReport(currentCheckResult []checker.Service, logPath string, cfg *configure.Configure) (reporter.Reporter, error) {
	// Load existing log data
	previousLog, err := common.ReadLogs(logPath)
	if err != nil {
		log.Printf("Error loading log data from %s: %v", logPath, err)
		return nil, err
	}

	// Use filtered data for further processing
	previousLog = common.FilterLogs(previousLog, currentCheckResult)

	// Extract service names in config order
	var serviceNames []string
	for _, service := range cfg.Services {
		serviceNames = append(serviceNames, service.Name)
	}

	// Parse log data into report format
	reportResult := reporter.ParseLogResult(previousLog, serviceNames, cfg)

	// calculate availability
	reportResult = getAvailability(reportResult)

	// calculate cert status
	reportResult = getCertStatus(reportResult, currentCheckResult)

	return reportResult, nil
}

// getAvailability calculates and updates the availability for each service in the report
func getAvailability(reportResult reporter.Reporter) reporter.Reporter {
	for i := range reportResult {
		if len(reportResult[i].ServiceHistory) == 0 {
			continue
		}
		statusAllEntryNum := 0
		for _, entry := range reportResult[i].ServiceHistory {
			if chk_result.IsALL(entry.Status) {
				statusAllEntryNum++
			}
		}
		availability := float64(statusAllEntryNum) / float64(len(reportResult[i].ServiceHistory))
		reportResult[i].Availability = availability
	}

	return reportResult
}

// getCertStatus updates the report with certificate status from the current check results
func getCertStatus(reportResult reporter.Reporter, currentCheckResult []checker.Service) reporter.Reporter {
	for _, serviceResult := range currentCheckResult {
		serviceName := serviceResult.Name

		// Find the service in the ordered slice
		for i := range reportResult {
			if reportResult[i].Name == serviceName {
				for _, endpointResult := range serviceResult.Endpoints {
					url := endpointResult.URL

					// Find the endpoint in the ordered slice and update it
					for j := range reportResult[i].Endpoints {
						if reportResult[i].Endpoints[j].URL == url {
							reportResult[i].Endpoints[j].IsHTTPS = endpointResult.IsHTTPS
							reportResult[i].Endpoints[j].CertRemainingDays = endpointResult.CertRemainingDays
							reportResult[i].Endpoints[j].IsCertExpired = endpointResult.IsCertExpired
							reportResult[i].Endpoints[j].DisplayURL = endpointResult.DisplayURL
							reportResult[i].Endpoints[j].HighlightSegments = endpointResult.HighlightSegments
							break
						}
					}
				}
				break
			}
		}
	}

	return reportResult
}

// WriteReport generates an HTML report from the provided log data
func WriteReport(reportResult reporter.Reporter, reportPath string, displayNum int) error {
	// Parse the HTML template
	tmpl, err := template.New("report.html").
		Funcs(createTemplateFunc()).
		ParseFiles(default_config.GetTemplatePath())
	if err != nil {
		return fmt.Errorf("template parsing failed: %w", err)
	}

	// Create or truncate the report file
	reportFile, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("file creation failed: %w", err)
	}
	defer func(reportFile *os.File) {
		if err := reportFile.Close(); err != nil {
			fmt.Printf("Error closing output file: %v\n", err)
		}
	}(reportFile)

	// Execute the template with the log data
	if err := tmpl.Execute(reportFile, map[string]any{
		"ReportResult": reportResult,
		"UpdateTime":   getLatestTime(reportResult),
		"DisplayNum":   displayNum,
	}); err != nil {
		return fmt.Errorf("template execution failed: %w", err)
	}

	return nil
}

// getLatestTime retrieves the latest time from the log data
func getLatestTime(reportResult reporter.Reporter) string {
	var latestTime string

	for _, serviceResult := range reportResult {
		for _, serviceHistoryEntry := range serviceResult.ServiceHistory {
			if latestTime == "" {
				latestTime = serviceHistoryEntry.Time
			} else if serviceHistoryEntry.Time > latestTime {
				latestTime = serviceHistoryEntry.Time
			}
		}
	}

	return latestTime
}

// createTemplateFunc defines custom template functions for the report
func createTemplateFunc() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b float64) float64 { return a * b },
		"div": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
		"until": func(n int) []int {
			result := make([]int, n)
			for i := range n {
				result[i] = i
			}
			return result
		},
	}
}
