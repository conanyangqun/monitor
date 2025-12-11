package checker

import (
	"log"
	"os"
	"strings"

	"github.com/wcy-dt/ponghub/internal/types/types/chk_result"
)

// isTestMode checks if the current execution is in test mode
func isTestMode() bool {
	// Check if any command line arguments contain "test"
	for _, arg := range os.Args {
		// Check for Go test flags (e.g., -test.v, -test.run, etc.)
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}

	return false
}

// logIfTest logs the message only if we're in test mode
func logIfTest(format string, args ...interface{}) {
	if isTestMode() {
		log.Printf(format, args...)
	}
}

// getTestResult determines the test result based on the success count and actual attempts
func getTestResult(successNum, attemptNum int) chk_result.CheckResult {
	switch successNum {
	case attemptNum:
		return chk_result.ALL
	case 0:
		return chk_result.NONE
	default:
		return chk_result.PART
	}
}
