package v2

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Validate ensures that all the time windows in t can be parsed.
func (t *TimeWindowWhen) Validate() error {
	if t == nil {
		return nil
	}
	for _, windows := range t.MapTimeWindows() {
		for _, window := range windows {
			if err := window.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Validate ensures the TimeWindowTimeRange is valid.
func (t *TimeWindowTimeRange) Validate() error {
	_, err := t.InWindow(time.Now())
	return err
}

// MapTimeWindows returns a map of all the time windows in t.
func (t *TimeWindowWhen) MapTimeWindows() map[string][]*TimeWindowTimeRange {
	d := t.Days
	return map[string][]*TimeWindowTimeRange{
		"All":       d.All,
		"Sunday":    d.Sunday,
		"Monday":    d.Monday,
		"Tuesday":   d.Tuesday,
		"Wednesday": d.Wednesday,
		"Thursday":  d.Thursday,
		"Friday":    d.Friday,
		"Saturday":  d.Saturday,
	}
}

// InWindow determines if the current time falls between the provided time
// window. Current should typically be time.Now() but to allow easier tests, it
// must be provided as a parameter. Begin and end parameters must be strings
// representing an hour of the day in the time.Kitchen format (e.g. "3:04PM")
func (t *TimeWindowTimeRange) InWindow(current time.Time) (bool, error) {
	// Get the year, month and day of the provided current time (e.g. 2016, 01 &
	// 02)
	year, month, day := current.Date()

	// Remove any whitespaces in the begin and end times, for backward
	// compatibility with Sensu v1 so "3:00 PM" becomes "3:00PM" and satisfies the
	// time.Kitchen format
	begin := strings.Replace(t.Begin, " ", "", -1)
	end := strings.Replace(t.End, " ", "", -1)

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

	return (current.After(beginTime) || current.Equal(beginTime)) &&
		(current.Before(endTime) || current.Equal(endTime)), nil
}

// InWindows determines if the current time falls between the provided time
// windows. Current should typically be time.Now() but to allow easier tests, it
// must be provided as a parameter. The function returns a positive value as
// soon the current time falls within a time window
func (t *TimeWindowWhen) InWindows(current time.Time) (bool, error) {
	windowsByDay := t.MapTimeWindows()

	var windows []*TimeWindowTimeRange
	windows = append(windows, windowsByDay["All"]...)
	windows = append(windows, windowsByDay[current.Weekday().String()]...)

	// Go through the set of matching windows and process all the individual
	// time windows. If the current time is within a time window in the selection,
	// then the loop returns early with true and nil error.
	for _, window := range windows {
		// Determine if the current time falls between this specific time window
		isInWindow, err := window.InWindow(current)
		if err != nil {
			return false, err
		}

		// Immediately return with a positive value if this time window conditions are
		// met
		if isInWindow {
			return true, nil
		}
	}

	// At this point no time windows conditions were met, return a negative value
	return false, nil
}

const (
	RepeatPeriodAnnually   = "annually"
	RepeatPeriodMonthly    = "monthly"
	RepeatPeriodWeekly     = "weekly"
	RepeatPeriodWeekdays   = "weekdays"
	RepeatPeriodWeekends   = "weekends"
	RepeatPeriodDaily      = "daily"
	RepeatPeriodSundays    = "sundays"
	RepeatPeriodMondays    = "mondays"
	RepeatPeriodTuesdays   = "tuesdays"
	RepeatPeriodWednesdays = "wednesdays"
	RepeatPeriodThursdays  = "thursdays"
	RepeatPeriodFridays    = "fridays"
	RepeatPeriodSaturdays  = "saturdays"
)

var validRepeatPeriods = map[string]bool{
	RepeatPeriodAnnually:   true,
	RepeatPeriodMonthly:    true,
	RepeatPeriodWeekly:     true,
	RepeatPeriodWeekdays:   true,
	RepeatPeriodWeekends:   true,
	RepeatPeriodDaily:      true,
	RepeatPeriodSundays:    true,
	RepeatPeriodMondays:    true,
	RepeatPeriodTuesdays:   true,
	RepeatPeriodWednesdays: true,
	RepeatPeriodThursdays:  true,
	RepeatPeriodFridays:    true,
	RepeatPeriodSaturdays:  true,
}

func (t *TimeWindowRepeated) Validate() error {
	begin, err := t.GetBeginTime()
	if err != nil {
		return fmt.Errorf("invalid begin time format: %v", err)
	}
	end, err := t.GetEndTime()
	if err != nil {
		return fmt.Errorf("invalid end time format: %v", err)
	}
	if end.Before(begin) {
		return errors.New("end time must be after begin time")
	}

	for _, repeat := range t.Repeat {
		if _, ok := validRepeatPeriods[repeat]; !ok {
			return fmt.Errorf("invalid repeat period: %s", repeat)
		}
	}

	return nil
}

func (t *TimeWindowRepeated) GetBeginTime() (time.Time, error) {
	beginTime, err := time.ParseInLocation(time.RFC3339, t.Begin, time.UTC)
	if err != nil {
		return time.Time{}, err
	}

	return beginTime, nil
}

func (t *TimeWindowRepeated) GetEndTime() (time.Time, error) {
	endTime, err := time.ParseInLocation(time.RFC3339, t.End, time.UTC)
	if err != nil {
		return time.Time{}, err
	}
	return endTime, nil
}

func (t *TimeWindowRepeated) InWindows(currentTime time.Time) bool {

	if len(t.Repeat) == 0 {
		return t.inAbsoluteTimeRange(currentTime)
	}

	for _, repeat := range t.Repeat {
		inTimeRange := false
		switch repeat {
		case RepeatPeriodAnnually:
			inTimeRange = t.inYearlyTimeRange(currentTime)
		case RepeatPeriodMonthly:
			inTimeRange = t.inMonthlyTimeRange(currentTime)
		case RepeatPeriodWeekly:
			inTimeRange = t.inWeeklyTimeRange(currentTime)
		case RepeatPeriodSundays:
			inTimeRange = t.inDayTimeRange(currentTime, time.Sunday)
		case RepeatPeriodMondays:
			inTimeRange = t.inDayTimeRange(currentTime, time.Monday)
		case RepeatPeriodTuesdays:
			inTimeRange = t.inDayTimeRange(currentTime, time.Tuesday)
		case RepeatPeriodWednesdays:
			inTimeRange = t.inDayTimeRange(currentTime, time.Wednesday)
		case RepeatPeriodThursdays:
			inTimeRange = t.inDayTimeRange(currentTime, time.Thursday)
		case RepeatPeriodFridays:
			inTimeRange = t.inDayTimeRange(currentTime, time.Friday)
		case RepeatPeriodSaturdays:
			inTimeRange = t.inDayTimeRange(currentTime, time.Saturday)
		case RepeatPeriodDaily:
			inTimeRange = t.inTimeRange(currentTime)
		case RepeatPeriodWeekdays:
			inTimeRange = t.inWeekdayTimeRange(currentTime)
		case RepeatPeriodWeekends:
			inTimeRange = t.inWeekendTimeRange(currentTime)
		}

		if inTimeRange {
			return inTimeRange
		}
	}

	return false
}

func (t *TimeWindowRepeated) inAbsoluteTimeRange(actualTime time.Time) bool {
	beginTime, err := t.GetBeginTime()
	if err != nil {
		return false
	}
	endTime, err := t.GetEndTime()
	if err != nil {
		return false
	}

	return actualTime.After(beginTime) && actualTime.Before(endTime)
}

func (t *TimeWindowRepeated) inDayTimeRange(actualTime time.Time, weekday time.Weekday) bool {
	beginTime, err := t.GetBeginTime()
	if err != nil {
		return false
	}
	endTime, err := t.GetEndTime()
	if err != nil {
		return false
	}

	if actualTime.Before(beginTime) {
		return false
	}

	actualTime = actualTime.In(beginTime.Location())
	duration := endTime.Sub(beginTime)
	beginHour, beginMin, beginSec := beginTime.Clock()

	thisWeekBegin := time.Date(actualTime.Year(), actualTime.Month(), actualTime.Day(), beginHour, beginMin, beginSec, 0, beginTime.Location())
	dayOffset := int(weekday) - int(actualTime.Weekday())
	thisWeekBegin = thisWeekBegin.AddDate(0, 0, dayOffset)

	thisWeekEnd := thisWeekBegin.Add(duration)
	thisWeekEnd.In(beginTime.Location())

	return actualTime.After(thisWeekBegin) && actualTime.Before(thisWeekEnd)
}

func (t *TimeWindowRepeated) inTimeRange(actualTime time.Time) bool {
	beginTime, err := t.GetBeginTime()
	if err != nil {
		return false
	}
	endTime, err := t.GetEndTime()
	if err != nil {
		return false
	}

	actualTime = actualTime.In(beginTime.Location())
	if actualTime.Before(beginTime) {
		return false
	}

	duration := endTime.Sub(beginTime)
	beginHour, beginMin, beginSec := beginTime.Clock()

	todayBegin := time.Date(actualTime.Year(), actualTime.Month(), actualTime.Day(), beginHour, beginMin, beginSec, 0, beginTime.Location())
	todayEnd := todayBegin.Add(duration)
	todayEnd = todayEnd.In(beginTime.Location())

	return actualTime.After(todayBegin) && actualTime.Before(todayEnd)
}

func (t *TimeWindowRepeated) inWeekdayTimeRange(actualTime time.Time) bool {
	return t.inDayTimeRange(actualTime, time.Monday) ||
		t.inDayTimeRange(actualTime, time.Tuesday) ||
		t.inDayTimeRange(actualTime, time.Wednesday) ||
		t.inDayTimeRange(actualTime, time.Thursday) ||
		t.inDayTimeRange(actualTime, time.Friday)
}

func (t *TimeWindowRepeated) inWeekendTimeRange(actualTime time.Time) bool {
	return t.inDayTimeRange(actualTime, time.Saturday) ||
		t.inDayTimeRange(actualTime, time.Sunday)
}

func (t *TimeWindowRepeated) inWeeklyTimeRange(actualTime time.Time) bool {
	beginTime, err := t.GetBeginTime()
	if err != nil {
		return false
	}

	return t.inDayTimeRange(actualTime, beginTime.Weekday())
}

func (t *TimeWindowRepeated) inMonthlyTimeRange(actualTime time.Time) bool {
	beginTime, err := t.GetBeginTime()
	if err != nil {
		return false
	}

	actualTime = actualTime.In(beginTime.Location())
	if beginTime.Day() != actualTime.Day() {
		return false
	}

	return t.inTimeRange(actualTime)
}

func (t *TimeWindowRepeated) inYearlyTimeRange(actualTime time.Time) bool {
	beginTime, err := t.GetBeginTime()
	if err != nil {
		return false
	}

	actualTime = actualTime.In(beginTime.Location())
	if beginTime.Day() != actualTime.Day() || beginTime.Month() != actualTime.Month() {
		return false
	}

	return t.inTimeRange(actualTime)
}
