package timeutil

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sensu/sensu-go/types"
)

const (
	kitchen24     = "15:04"
	kitchenOffset = "15:04 -07:00"
	legacy        = "Jan 02 2006 3:04PM"
	rfc3339Space  = "2006-01-02 15:04:05 Z07:00"
)

var (
	// kitchen12Re represents the time.Kitchen format: 3:04PM
	kitchen12Re = regexp.MustCompile(`^([0-1]?[0-9]:[0-5][0-9])\s?(AM|PM)( .+)?$`)

	// kitchen24Re represents the the kitchen format but in 24-hour format: 15:04
	kitchen24Re = regexp.MustCompile(`^([01][0-9]|2[0-3])(:?)([0-5][0-9])( .+)?$`)

	// legacyRe represents the legacy format used in Sensu 2 alpha releases: Jan
	// 02 2006 3:04PM MST
	legacyRe = regexp.MustCompile(`([A-Z][a-z]{2}) ` + // Month (i.e. May)
		`(0[1-9]|[1-2][0-9]|3[0-1]) ` + // Day (i.e. 14)
		`([0-9]{4}) ` + // Year (i.e. 2018)
		`([0-1]?[0-9]:[0-5][0-9](?:AM|PM))` + // Hour (i.e. 3:04PM)
		`( .+)?`) // Timezone (e.g. MST or America/New_York)

	// offsetTz represents a numeric zone offset, e.g. -07:00
	offsetTz = regexp.MustCompile(`^([+-](?:2[0-3]|[01][0-9]):[0-5][0-9])$`)

	// rfc3339Re represents the time.RFC3339 format
	rfc3339Re = regexp.MustCompile(`^(\d+)` + // year (i.e. 2018)
		`-` +
		`(0[1-9]|1[012])` + // month (i.e. 05)
		`-` +
		`(0[1-9]|[12]\d|3[01])` + // day (i.e. 14)
		`T` +
		`([01]\d|2[0-3])` + // hour (i.e. 15)
		`:` +
		`([0-5]\d)` + // minute (i.e. 04)
		`:` +
		`([0-5]\d|60)` + // second (i.e. 05)
		`(Z|([\+|\-]([01][0-9]|2[0-3]):[0-5][0-9]))$`) // zone (e.g. Z or -07:00)

	// rfc3339SpaceRe represents the time.RFC3339 format but with space delimiters
	rfc3339SpaceRe = regexp.MustCompile(`^(\d+)` + // year (i.e. 2018)
		`-` +
		`(0[1-9]|1[012])` + // month (i.e. 05)
		`-` +
		`(0[1-9]|[12]\d|3[01])` + // day (i.e. 14)
		`\s` +
		`([01]\d|2[0-3])` + // hour (i.e. 15)
		`:` +
		`([0-5]\d)` + // minute (i.e. 04)
		`:` +
		`([0-5]\d|60)` + // second (i.e. 05)
		`\s` +
		`(Z|([\+|\-]([01][0-9]|2[0-3]):[0-5][0-9]))$`) // zone (e.g. Z or -07:00)
)

// HumanTimestamp takes a timestamp and returns a readable date using the format
// string "2006-01-02 15:04:05.999999999 -0700 MST". If the timestamp equals 0,
// "N/A" will be returned instead of the epoch date
func HumanTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return "N/A"
	}

	return time.Unix(timestamp, 0).String()
}

// ConvertToUTC takes a TimeWindowRange and converts both the begin time and
// end time of the window to UTC
func ConvertToUTC(t *types.TimeWindowTimeRange) error {
	begin, err := kitchenToTime(t.Begin)
	if err != nil {
		return err
	}

	end, err := kitchenToTime(t.End)
	if err != nil {
		return err
	}

	t.Begin = begin.UTC().Format(time.Kitchen)
	t.End = end.UTC().Format(time.Kitchen)
	return nil
}

// ConvertToUnix takes a full date and converts it to a UNIX timestamp
func ConvertToUnix(value string) (int64, error) {
	if value == "0" || value == "now" {
		return time.Now().Unix(), nil
	}

	t, err := dateToTime(value)
	if err != nil {
		return 0, err
	}

	return t.Unix(), nil
}

// dateToTime takes a full date in an unknown format and tries to detect it
// using various formats. The time is returned as time.Time or an error if no
// matching format was found
func dateToTime(str string) (time.Time, error) {
	// Try RFC3339 format (2006-01-02T15:04:05-07:00)
	if match := rfc3339Re.FindString(str); match != "" {
		return time.Parse(time.RFC3339, match)
	}

	// Try RFC3339 format with space delimiter (2006-01-02 15:04:05 -07:00)
	if match := rfc3339SpaceRe.FindString(str); match != "" {
		return time.Parse(rfc3339Space, match)
	}

	// Try legacy format (Jan 02 2006 3:04PM UTC)
	if matches := legacyRe.FindStringSubmatch(str); len(matches) > 0 {
		// Extract the timezone
		tz := strings.TrimSpace(matches[5])

		// Reassemble the time, without any timezone
		t := strings.Join(matches[1:len(matches)-1], " ")

		return parseInLocaton(legacy, t, tz)
	}

	return time.Time{}, fmt.Errorf("unknown format for provided date %s", str)
}

// kitchenToTime takes a kitchen time (without a date) in an unknown format and
// tries to detect it using various formats. The time is returned as time.Time
// or an error if no matching format was found
func kitchenToTime(str string) (time.Time, error) {
	// Try 24-hour kitchen format (15:04)
	if matches := kitchen24Re.FindStringSubmatch(str); len(matches) > 0 {
		// Extract the timezone
		tz := strings.TrimSpace(matches[4])

		// Verify if we have a numerical zone offset (i.e. -07:00)
		if match := offsetTz.FindString(tz); match != "" {
			return time.Parse(kitchenOffset, matches[0])
		}

		// Reassemble the time, without any timezone
		t := strings.Join(matches[1:len(matches)-1], "")

		return parseInLocaton(kitchen24, t, tz)
	}

	// Try 12-hour kitchen format (3:04PM)
	if matches := kitchen12Re.FindStringSubmatch(str); len(matches) > 0 {
		// Extract the timezone
		tz := strings.TrimSpace(matches[3])

		// Reassemble the time, without any timezone
		t := strings.Join(matches[1:len(matches)-1], "")

		return parseInLocaton(time.Kitchen, t, tz)
	}

	return time.Time{}, fmt.Errorf("unknown format for provided time %s", str)
}

// extractLocation extracts the location from the time value, using the regex,
// and returns a standardized time value, along with its location and any error
// encountered
func extractLocation(location string) (*time.Location, error) {
	if location == "" {
		return time.Local, nil
	}

	return time.LoadLocation(location)
}

func parseInLocaton(layout, value, locString string) (time.Time, error) {
	loc, err := extractLocation(locString)
	if err != nil {
		return time.Time{}, err
	}

	return time.ParseInLocation(layout, value, loc)
}
