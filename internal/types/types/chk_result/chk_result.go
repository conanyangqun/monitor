package chk_result

type CheckResult string

const (
	// ALL represents all ports are online
	ALL CheckResult = "all"

	// PART represents some ports are online
	PART CheckResult = "part"

	// NONE represents no ports are online
	NONE CheckResult = "none"

	// UNKNOWN represents an unknown test result
	UNKNOWN CheckResult = "unknown"
)

// String returns the string representation of the CheckResult
func (tr CheckResult) String() string {
	switch tr {
	case ALL:
		return "all"
	case PART:
		return "part"
	case NONE:
		return "none"
	default:
		return "unknown"
	}
}

// IsValid checks if the CheckResult is valid
func (tr CheckResult) IsValid() bool {
	return tr == ALL || tr == PART || tr == NONE
}

// IsALL checks if the CheckResult is ALL
func IsALL(resultStr string) bool {
	return ParseCheckResult(resultStr) == ALL
}

// ParseCheckResult parses a string into a CheckResult
func ParseCheckResult(s string) CheckResult {
	switch s {
	case "all":
		return ALL
	case "part":
		return PART
	case "none":
		return NONE
	default:
		return UNKNOWN
	}
}
