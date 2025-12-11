package checker

import (
	"time"

	"github.com/wcy-dt/ponghub/internal/types/structures/checker"
	"github.com/wcy-dt/ponghub/internal/types/structures/configure"
	"github.com/wcy-dt/ponghub/internal/types/types/chk_result"
)

// CheckServices checks all services defined in the configuration
func CheckServices(cfg *configure.Configure) []checker.Service {
	var checkResult []checker.Service
	for _, service := range cfg.Services {
		attemptNum := 0
		successNum := 0
		endpointNum := 0
		onlineEndpointNum := 0

		// check Endpoints ports
		startTime := time.Now()
		var endpointResults []checker.Endpoint
		for _, endpoint := range service.Endpoints {
			endpointResult := checkEndpoint(&endpoint, service.Timeout, service.MaxRetryTimes, service.Name)
			endpointResults = append(endpointResults, endpointResult)
			attemptNum += endpointResult.AttemptNum
			successNum += endpointResult.SuccessNum
			endpointNum++
			if endpointResult.Status == chk_result.ALL {
				onlineEndpointNum++
			}
		}
		endTime := time.Now()

		serviceResult := checker.Service{
			Name:       service.Name,
			Status:     getTestResult(onlineEndpointNum, endpointNum),
			Endpoints:  endpointResults,
			StartTime:  startTime.Format(time.RFC3339),
			EndTime:    endTime.Format(time.RFC3339),
			AttemptNum: attemptNum,
			SuccessNum: successNum,
		}
		checkResult = append(checkResult, serviceResult)
	}
	return checkResult
}
