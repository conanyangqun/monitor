package checker

import (
	"time"

	"github.com/wcy-dt/ponghub/internal/types/types/chk_result"
	"github.com/wcy-dt/ponghub/internal/types/types/highlight"
)

// Result defines the structure for the result of checking a service
type (
	// Service defines the structure for the result of checking a service
	Service struct {
		Name       string                 `json:"name"`
		Status     chk_result.CheckResult `json:"status"`
		Endpoints  []Endpoint             `json:"endpoints,omitempty"`
		StartTime  string                 `json:"start_time"`
		EndTime    string                 `json:"end_time"`
		AttemptNum int                    `json:"attempt_num"`
		SuccessNum int                    `json:"success_num"`
	}

	// Endpoint defines the structure for the result of checking a port
	Endpoint struct {
		URL               string                 `json:"url"`
		Method            string                 `json:"method"`
		Body              string                 `json:"body,omitempty"`
		Status            chk_result.CheckResult `json:"status"`
		StatusCode        int                    `json:"status_code,omitempty"`
		StartTime         string                 `json:"start_time"`
		EndTime           string                 `json:"end_time"`
		ResponseTime      time.Duration          `json:"response_time"`
		AttemptNum        int                    `json:"attempt_num"`
		SuccessNum        int                    `json:"success_num"`
		FailureDetails    []string               `json:"failure_details,omitempty"`
		ResponseBody      string                 `json:"response_body,omitempty"`
		IsHTTPS           bool                   `json:"is_https,omitempty"`
		CertRemainingDays int                    `json:"cert_remaining_days,omitempty"`
		IsCertExpired     bool                   `json:"is_cert_expired,omitempty"`
		DisplayURL        string                 `json:"display_url,omitempty"`
		HighlightSegments []highlight.Segment    `json:"highlight_segments,omitempty"`
	}
)
