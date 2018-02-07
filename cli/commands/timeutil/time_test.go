package timeutil

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOffsetTime(t *testing.T) {
	tests := []struct {
		In          string
		Want        string
		ExpectedErr bool
		Format      string
		Regex       *regexp.Regexp
	}{
		{"8:00AM MST", "3:00PM", false, time.Kitchen, kitchenTZRE},
		{"800AM", "12:00AM", true, time.Kitchen, kitchenTZRE},
		{"Feb 06 2018 8:00AM MST", "Feb 06 2018 3:00PM", false, dateFormat, dateFormatTZRE},
		{"Feb 06 2018 8:00AM foo", "Jan 01 0001 12:00AM", true, dateFormat, dateFormatTZRE},
		{"foo", "Jan 01 0001 12:00AM", true, dateFormat, dateFormatTZRE},
		{"Feb 06 2018 800AM", "Jan 01 0001 12:00AM", true, dateFormat, dateFormatTZRE},
	}
	for _, test := range tests {
		t.Run(test.In, func(t *testing.T) {
			got, err := offsetTime(test.In, test.Format, test.Regex)
			if !test.ExpectedErr {
				require.NoError(t, err)
			}
			require.Equal(t, test.Want, got.Format(test.Format))
		})
	}
}
