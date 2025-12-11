package common

import (
	"time"

	"github.com/wcy-dt/ponghub/internal/types/structures/checker"
	"github.com/wcy-dt/ponghub/internal/types/types/chk_result"
)

// processCheckResult processes the check results for a service
func processCheckResult(serviceResult checker.Service) (map[string][]chk_result.CheckResult, map[string]string, map[string]time.Duration) {
	urlStatusMap := make(map[string][]chk_result.CheckResult)
	urlTimeMap := make(map[string]string)
	urlResponseTimeMap := make(map[string]time.Duration)

	// Process Endpoints checks
	for _, endpoint := range serviceResult.Endpoints {
		urlStatusMap[endpoint.URL] = append(urlStatusMap[endpoint.URL], endpoint.Status)

		if _, exists := urlTimeMap[endpoint.URL]; !exists {
			urlTimeMap[endpoint.URL] = endpoint.StartTime
		} else if endpoint.StartTime < urlTimeMap[endpoint.URL] {
			urlTimeMap[endpoint.URL] = endpoint.StartTime
		}

		if _, exists := urlResponseTimeMap[endpoint.URL]; !exists {
			urlResponseTimeMap[endpoint.URL] = endpoint.ResponseTime
		} else if endpoint.ResponseTime > urlResponseTimeMap[endpoint.URL] {
			urlResponseTimeMap[endpoint.URL] = endpoint.ResponseTime
		}
	}

	return urlStatusMap, urlTimeMap, urlResponseTimeMap
}
