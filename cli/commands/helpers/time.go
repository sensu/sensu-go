package helpers

import "time"

// HumanTimestamp takes a timestamp and returns a readable date. If the
// timestamp equals 0, "N/A" will be returned instead of the epoch date
func HumanTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return "N/A"
	}

	return time.Unix(timestamp, 0).String()
}
