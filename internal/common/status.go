package common

import "github.com/wcy-dt/ponghub/internal/types/types/chk_result"

// calcMergedStatus merges multiple statuses into a single status
func calcMergedStatus(statusList []chk_result.CheckResult) chk_result.CheckResult {
	if len(statusList) == 0 {
		return chk_result.NONE
	}

	hasNone, hasAll := false, false
	for _, s := range statusList {
		switch s {
		case chk_result.NONE:
			hasNone = true
		case chk_result.ALL:
			hasAll = true
		}
	}

	switch {
	case hasNone && !hasAll:
		return chk_result.NONE
	case !hasNone && hasAll:
		return chk_result.ALL
	default:
		return chk_result.PART
	}
}
