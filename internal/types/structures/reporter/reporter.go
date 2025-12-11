package reporter

import "github.com/wcy-dt/ponghub/internal/types/types/highlight"

// Data structures for logging and reporting
type (
	// HistoryEntry represents a single history entry
	HistoryEntry struct {
		Time         string
		Status       string
		ResponseTime int
	}

	History []HistoryEntry

	Endpoint struct {
		URL               string // Added URL field to store the endpoint URL
		EndpointHistory   History
		IsHTTPS           bool
		IsCertExpired     bool
		CertRemainingDays int
		DisplayURL        string              // Resolved URL for display
		HighlightSegments []highlight.Segment // Segments with highlight info
	}

	// Endpoints is a slice of Endpoint
	Endpoints []Endpoint

	// Service represents the result of checking a service
	Service struct {
		Name           string // Added Name field to identify the service
		ServiceHistory History
		Availability   float64
		Endpoints      Endpoints
	}

	// Reporter is a slice of Service
	Reporter []Service
)
