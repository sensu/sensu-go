package timeutil

import (
	"regexp"
	"strings"
	"time"

	"github.com/sensu/sensu-go/types"
)

const dateFormat = "Jan 02 2006 3:04PM"

var kitchenTZRE = regexp.MustCompile(`[0-1]?[0-9]:[0-5][0-9]\s?(AM|PM)( .+)?`)
var dateFormatTZRE = regexp.MustCompile(`[A-Z][a-z]{2} (0[1-9]|[1-2][0-9]|3[0-1]) [0-9]{4} [0-1]?[0-9]:[0-5][0-9](AM|PM)( .+)?`)

// HumanTimestamp takes a timestamp and returns a readable date. If the
// timestamp equals 0, "N/A" will be returned instead of the epoch date
func HumanTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return "N/A"
	}

	return time.Unix(timestamp, 0).String()
}

// ConvertToUTC takes a TimeWindowRange and converts both the begin time and
// end time of the window to UTC
func ConvertToUTC(t *types.TimeWindowTimeRange) error {
	begin, err := offsetTime(t.Begin, time.Kitchen, kitchenTZRE)
	if err != nil {
		return nil
	}
	end, err := offsetTime(t.End, time.Kitchen, kitchenTZRE)
	if err != nil {
		return nil
	}
	t.Begin = begin.Format(time.Kitchen)
	t.End = end.Format(time.Kitchen)
	return nil
}

// ConvertToUnixUTC takes a string formatted as dateFormat and converts it to
// UTC in unix epoch form
func ConvertToUnixUTC(begin string) (int64, error) {
	if begin == "0" {
		return 0, nil
	}
	utc, err := offsetTime(begin, dateFormat, dateFormatTZRE)
	if err != nil {
		return 0, err
	}
	return utc.Unix(), nil
}

func offsetTime(s string, fs string, format *regexp.Regexp) (time.Time, error) {
	ts, tz, err := extractLocation(s, format)
	if err != nil {
		return time.Time{}, err
	}
	tm, err := time.ParseInLocation(fs, ts, tz)
	if err != nil {
		return time.Time{}, err
	}
	_, offset := tm.Zone()
	tm = tm.Add(-time.Duration(offset) * time.Second)
	return tm, nil
}

func extractLocation(s string, format *regexp.Regexp) (string, *time.Location, error) {
	tz := time.Local
	beginMatches := format.FindStringSubmatch(s)
	if len(beginMatches) == 0 {
		return s, tz, nil
	}
	possibleTZ := strings.TrimSpace(beginMatches[len(beginMatches)-1])
	if len(possibleTZ) == 0 {
		return s, tz, nil
	}
	loc, err := time.LoadLocation(possibleTZ)
	trimmed := strings.TrimSpace(strings.TrimSuffix(s, possibleTZ))
	normalized := strings.Replace(strings.Replace(trimmed, " AM", "AM", -1), " PM", "PM", -1)
	return normalized, loc, err
}
