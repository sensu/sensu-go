package time

import "time"

// InWindow determines if the current time falls between the provided time
// window. Current should typically be time.Now() but to allow easier tests, it
// must be provided as a parameter. Begin and end parameters must be strings
// representing an hour of the day in the time.Kitchen format (e.g. "3:04PM")
func InWindow(current time.Time, begin, end string) (bool, error) {
	// Get the year, month and day of the provided current time (e.g. 2016, 01 &
	// 02)
	year, month, day := current.Date()

	// Parse the beginning of the provided time window in order to retrieve the
	// hour and minute and apply it to current year, month and day so we end up
	// with a date that corresponds to today (e.g. 2006-01-02T15:00:00Z)
	beginTime, err := time.Parse(time.Kitchen, begin)
	if err != nil {
		return false, err
	}
	beginHour, beginMin, _ := beginTime.Clock()
	beginTime = time.Date(year, month, day, beginHour, beginMin, 0, 0, time.UTC)

	// Parse the ending of the provided time window in order to retrieve the
	// hour and minute and apply it to current year, month and day so we end up
	// with a date that corresponds to today (e.g. 2006-01-02T21:00:00Z)
	endTime, err := time.Parse(time.Kitchen, end)
	if err != nil {
		return false, err
	}
	endHour, endMin, _ := endTime.Clock()
	endTime = time.Date(year, month, day, endHour, endMin, 0, 0, time.UTC)

	// Verify if the end of the time window is actually before the beginning of
	// it, which means that the window ends the next day (e.g. 3:00PM to 8:00AM)
	if endTime.Before(beginTime) {
		// Verify if the current time is before the end of the time window, which
		// means that we are already on the second day of the specified time window,
		// therefore we just need to move the start of this window to the beginning
		// of this second day (e.g. 3:00PM to 8:00AM, it's currently 5:00AM so let's
		// move the beginning to 0:00AM)
		if current.Before(endTime) {
			beginTime = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		} else {
			// We are currently on the first day of the window so we just need to move
			// the end of this window to the end of the first day (e.g. 3:00PM to
			// 8:00AM, it's currently 5:00PM so let's move the ending to 11:59PM)
			endTime = time.Date(year, month, day, 23, 59, 59, 999999999, time.UTC)
		}
	}

	return current.After(beginTime) && current.Before(endTime), nil
}
