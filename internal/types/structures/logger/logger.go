package logger

type (
	// HistoryEntry represents a single history entry
	HistoryEntry struct {
		Time         string `json:"time"`
		Status       string `json:"status"`
		ResponseTime int    `json:"response_time,omitempty"`
	}

	History   []HistoryEntry
	Endpoints map[string]History

	// Service represents log data for a service
	Service struct {
		ServiceHistory History   `json:"service_history"`
		Endpoints      Endpoints `json:"endpoints"`
	}

	// Logger represents the entire log structure
	Logger map[string]Service
)
