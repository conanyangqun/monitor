package configure

type (
	// Configure defines the overall configuration structure for the application
	Configure struct {
		Services       []Service           `yaml:"services"`
		Timeout        int                 `yaml:"timeout,omitempty"`
		MaxRetryTimes  int                 `yaml:"max_retry_times,omitempty"`
		MaxLogDays     int                 `yaml:"max_log_days,omitempty"`
		CertNotifyDays int                 `yaml:"cert_notify_days,omitempty"`
		DisplayNum     int                 `yaml:"display_num,omitempty"`
		Notifications  *NotificationConfig `yaml:"notifications,omitempty"`
	}
)
