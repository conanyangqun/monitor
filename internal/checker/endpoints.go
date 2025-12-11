package checker

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/wcy-dt/ponghub/internal/common/params"
	"github.com/wcy-dt/ponghub/internal/types/structures/checker"
	"github.com/wcy-dt/ponghub/internal/types/structures/configure"
)

// checkEndpoint checks a single port based on the provided configuration
func checkEndpoint(cfg *configure.Endpoint, timeout int, maxRetryTimes int, serviceName string) checker.Endpoint {
	var failureDetails []string
	successNum := 0
	attemptNum := 0

	var statusCode int
	var responseBody string

	httpMethod := getHttpMethod(cfg.Method)
	maxResponseTime := time.Duration(0)

	// SSL certificate related variables
	urlIsHTTPS := isHTTPS(cfg.ParsedURL)
	certRemainingDays := 0
	isCertExpired := false

	// Generate display URL for smart showing of template vs resolved URL
	resolver := params.NewParameterResolver()
	displayURL, highlightSegments := resolver.HighlightChanges(cfg.URL)
	originalURL := cfg.URL
	if originalURL == "" {
		originalURL = cfg.ParsedURL
		displayURL = cfg.ParsedURL
		highlightSegments = nil
	}

	// Check SSL certificate if it's an HTTPS URL
	if urlIsHTTPS {
		remainingDays, expired, err := checkSSLCertificates(cfg.ParsedURL)
		if err != nil {
			urlIsHTTPS = false
			// Only log success details during tests to avoid exposing secrets
			logIfTest("SSL certificate check failed for %s: %v", cfg.ParsedURL, err)
			failureDetails = append(failureDetails, fmt.Sprintf("SSL Certificate Error: %s", err.Error()))
		} else {
			certRemainingDays = remainingDays
			isCertExpired = expired
			// Only log success details during tests to avoid exposing secrets
			logIfTest("SSL Certificate Info for %s: %d days remaining, expired: %v", cfg.ParsedURL, remainingDays, expired)
		}
	}

	startTime := time.Now()
	for currentAttemptNum := range maxRetryTimes {
		attemptNum++
		client := &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
		// Only log request details during tests to avoid exposing secrets
		logIfTest("[%s] %s %s (attempt %d/%d)\n",
			serviceName, httpMethod, cfg.ParsedURL, currentAttemptNum+1, maxRetryTimes)

		// build the request
		req, err := http.NewRequest(httpMethod, cfg.ParsedURL, nil)
		if err != nil {
			failureDetails = append(failureDetails, fmt.Sprintf("StatusCode: N/A, Error: %s", err.Error()))
			log.Printf("FAILED - Error: %s", err.Error())
			continue
		}
		for headerName, headerValue := range cfg.ParsedHeaders {
			req.Header.Set(headerName, headerValue)
		}
		if cfg.ParsedBody != "" {
			req.Body = io.NopCloser(strings.NewReader(cfg.ParsedBody))
		}

		// get the response
		reqStartTime := time.Now()
		resp, err := client.Do(req)
		responseTime := time.Since(reqStartTime)
		if err != nil {
			failureDetails = append(failureDetails, fmt.Sprintf("StatusCode: N/A, Error: %s", err.Error()))
			log.Printf("FAILED - Error: %s", err.Error())
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			failureDetails = append(failureDetails, fmt.Sprintf("StatusCode: %d, Error: %s", resp.StatusCode, err.Error()))
			log.Printf("FAILED - StatusCode: %d, Error: %s", resp.StatusCode, err.Error())
			if err := resp.Body.Close(); err != nil {
				// Only log response body errors during tests to avoid exposing secrets
				logIfTest("Error closing response body for %s: %v", cfg.ParsedURL, err)
			}
			continue
		}
		responseBody = string(body)
		statusCode = resp.StatusCode

		// check the response
		isOnline := isSuccessfulResponse(cfg, resp, body)
		if isOnline {
			successNum++
			if responseTime > maxResponseTime {
				maxResponseTime = responseTime
			}
			responseBody = ""
			if err := resp.Body.Close(); err != nil {
				// Only log response body errors during tests to avoid exposing secrets
				logIfTest("Error closing response body for %s: %v", cfg.ParsedURL, err)
			}
			// Only log success details during tests to avoid exposing secrets
			logIfTest("SUCCESS - %s %s (attempt %d/%d) - Response Time: %d ms, Status Code: %d",
				httpMethod, cfg.ParsedURL, currentAttemptNum+1, maxRetryTimes, responseTime.Milliseconds(), resp.StatusCode)
			break
		}
		failureDetails = append(failureDetails, fmt.Sprintf("StatusCode or ResponseRegex mismatch: %d", resp.StatusCode))
		log.Printf("FAILED - StatusCode or ResponseRegex mismatch: %d", resp.StatusCode)
		if err := resp.Body.Close(); err != nil {
			// Only log response body errors during tests to avoid exposing secrets
			logIfTest("Error closing response body for %s: %v", cfg.ParsedURL, err)
		}
	}
	endTime := time.Now()

	return checker.Endpoint{
		URL:               cfg.URL,
		Method:            httpMethod,
		Body:              cfg.Body,
		Status:            getTestResult(successNum, attemptNum),
		StatusCode:        statusCode,
		StartTime:         startTime.Format(time.RFC3339),
		EndTime:           endTime.Format(time.RFC3339),
		ResponseTime:      maxResponseTime,
		AttemptNum:        attemptNum,
		SuccessNum:        successNum,
		FailureDetails:    failureDetails,
		ResponseBody:      responseBody,
		IsHTTPS:           urlIsHTTPS,
		CertRemainingDays: certRemainingDays,
		IsCertExpired:     isCertExpired,
		DisplayURL:        displayURL,
		HighlightSegments: highlightSegments,
	}
}

// getHttpMethod converts a string method to an HTTP method constant
func getHttpMethod(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return http.MethodGet
	case "POST":
		return http.MethodPost
	case "PUT":
		return http.MethodPut
	case "DELETE":
		log.Fatalln(errors.New("method not supported"))
	case "HEAD":
		log.Fatalln(errors.New("method not supported"))
	case "PATCH":
		log.Fatalln(errors.New("method not supported"))
	case "OPTIONS":
		log.Fatalln(errors.New("method not supported"))
	case "TRACE":
		log.Fatalln(errors.New("method not supported"))
	case "CONNECT":
		log.Fatalln(errors.New("method not supported"))
	default:
		return http.MethodGet // Default to GET if method is unknown
	}
	return http.MethodGet
}

// isSuccessfulResponse checks if the response from the server is successful based on the configuration
func isSuccessfulResponse(cfg *configure.Endpoint, rsp *http.Response, body []byte) bool {
	// responseRegex is set, and the response body does not match the regex
	if cfg.ResponseRegex != "" {
		matched, err := regexp.Match(cfg.ResponseRegex, body)
		if err != nil {
			log.Fatalln("Error parsing regexp:", err)
		}
		if !matched {
			return false
		}
	}

	// statusCode and responseRegex are not set, and the response is OK
	if cfg.StatusCode == 0 && cfg.ResponseRegex == "" && rsp.StatusCode == http.StatusOK {
		return true
	}

	// statusCode is not set, and the responseRegex matches
	if cfg.StatusCode == 0 && cfg.ResponseRegex != "" {
		return true
	}

	// statusCode is set, and the response matches the expected status code
	if cfg.StatusCode != 0 && rsp.StatusCode == cfg.StatusCode {
		return true
	}

	return false
}
