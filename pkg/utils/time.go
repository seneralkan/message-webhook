package utils

import "time"

// GetCurrentTimestamp returns the current timestamp in nanoseconds
func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}
