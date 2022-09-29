package metrics

import "time"

func equalStrings(s1 []string, s2 []string) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if len(s1) != len(s2) {
		return false
	}
	for i, val := range s1 {
		if s2[i] != val {
			return false
		}
	}
	return true
}

// nowFunc allows altering the result of Now for testing
var nowFunc func() time.Time

func init() {
	nowFunc = time.Now
}

// SetUpNowTime overrides the 'now' time seen by this library to a constant value (instead of time.Now)
// This function is provided only to ease unit testing
func SetUpNowTime(t time.Time) {
	nowFunc = func() time.Time {
		return t
	}
}

// RestoreNowTime restores the 'now' time seen by this library to time.Now.
func RestoreNowTime() {
	nowFunc = time.Now
}
