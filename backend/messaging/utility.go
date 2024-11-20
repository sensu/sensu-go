package messaging

// utility.go

import (
	"log"
	"time"
)

// convertToLocalTime converts a given time to the local timezone.
func convertToLocalTime(t time.Time) time.Time {
	// Load the local timezone location (this could be adjusted based on user preferences)
	location, err := time.LoadLocation("Local") // "Local" refers to the local timezone
	if err != nil {
		log.Printf("Error loading local timezone: %v, using UTC instead", err)
		// If loading local timezone fails, fallback to UTC
		location = time.UTC
	}
	return t.In(location)
}
