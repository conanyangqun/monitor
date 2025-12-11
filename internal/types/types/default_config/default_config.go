package default_config

const (
	// timeout is the default timeout for service checks in seconds
	timeout = 5

	// maxRetryTimes is the default maxRetryTimes count for service checks
	maxRetryTimes = 2

	// maxLogDays is the default maximum number of days to keep logs
	maxLogDays = 3

	// certNotifyDays is the default number of days to notify before certificate expiration
	certNotifyDays = 7
)

// GetDefaultTimeout returns the default timeout for service checks
func GetDefaultTimeout() int {
	return timeout
}

// GetDefaultMaxRetryTimes returns the default maxRetryTimes count for service checks
func GetDefaultMaxRetryTimes() int {
	return maxRetryTimes
}

// GetDefaultMaxLogDays returns the default maximum number of days to keep logs
func GetDefaultMaxLogDays() int {
	return maxLogDays
}

// GetDefaultCertNotifyDays returns the default number of days to notify before certificate expiration
func GetDefaultCertNotifyDays() int {
	return certNotifyDays
}

// SetDefaultTimeout sets the default timeout for a given configuration pointer
func SetDefaultTimeout(cfg *int) {
	if *cfg <= 0 {
		*cfg = GetDefaultTimeout()
	}
}

// SetDefaultMaxRetryTimes sets the default maxRetryTimes count for a given configuration pointer
func SetDefaultMaxRetryTimes(cfg *int) {
	if *cfg <= 0 {
		*cfg = GetDefaultMaxRetryTimes()
	}
}

// SetDefaultMaxLogDays sets the default maximum number of days to keep logs for a given configuration pointer
func SetDefaultMaxLogDays(cfg *int) {
	if *cfg <= 0 {
		*cfg = GetDefaultMaxLogDays()
	}
}

// SetDefaultCertNotifyDays sets the default number of days to notify before certificate expiration for a given configuration pointer
func SetDefaultCertNotifyDays(cfg *int) {
	if *cfg <= 0 {
		*cfg = GetDefaultCertNotifyDays()
	}
}

const (
	// displayNum is the default number of logs per endpoint to display in the HTML report
	displayNum = 72
)

// GetDisplayNum returns the default number of logs per endpoint to display in the HTML report
func GetDisplayNum() int {
	return displayNum
}

// SetDefaultDisplayNum sets the default number of logs per endpoint to display in the HTML report for a given configuration pointer
func SetDefaultDisplayNum(cfg *int) {
	if *cfg <= 0 {
		*cfg = GetDisplayNum()
	}
}

const (
	// configPath is the default path to the configuration file
	configPath = "config.yaml"

	// logPath is the default path to the data file where logs are stored
	logPath = "data/ponghub_log.json"

	// reportPath is the default path to the HTML report file
	reportPath = "data/index.html"

	// templatePath is the default path to the HTML template file
	templatePath = "templates/report.html"

	// notifyPath is the default path to the notification template file
	notifyPath = "data/notify.txt"
)

// GetConfigPath returns the default path to the configuration file
func GetConfigPath() string {
	return configPath
}

// GetLogPath returns the default path to the data file where logs are stored
func GetLogPath() string {
	return logPath
}

// GetReportPath returns the default path to the HTML report file
func GetReportPath() string {
	return reportPath
}

// GetTemplatePath returns the default path to the HTML template file
func GetTemplatePath() string {
	return templatePath
}

// GetNotifyPath returns the default path to the notification template file
func GetNotifyPath() string {
	return notifyPath
}
