package v2

import (
	"testing"
	time "time"

	"github.com/stretchr/testify/assert"
)

func mustParse(t *testing.T, timeStr string) time.Time {
	tm, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

func TestInWindow(t *testing.T) {
	testCases := []struct {
		name          string
		now           time.Time
		window        TimeWindowTimeRange
		expected      bool
		expectedError bool
	}{
		{
			name: "is within window",
			now:  mustParse(t, "2006-01-02T15:04:05Z"),
			window: TimeWindowTimeRange{
				Begin: "3:00PM",
				End:   "4:00PM",
			},
			expected:      true,
			expectedError: false,
		},
		{
			name: "is outside window",
			now:  mustParse(t, "2006-01-02T10:04:05Z"),
			window: TimeWindowTimeRange{
				Begin: "3:00PM",
				End:   "4:00PM",
			},
			expected:      false,
			expectedError: false,
		},
		{
			name: "unsupported time window beginning format",
			now:  mustParse(t, "2006-01-02T10:04:05Z"),
			window: TimeWindowTimeRange{
				Begin: "15:00",
				End:   "4:00PM",
			},
			expected:      false,
			expectedError: true,
		},
		{
			name: "unsupported time window ending format",
			now:  mustParse(t, "2006-01-02T10:04:05Z"),
			window: TimeWindowTimeRange{
				Begin: "3:00PM",
				End:   "16:00",
			},
			expected:      false,
			expectedError: true,
		},
		{
			name: "supports time window with whitespaces for backward compatibility",
			now:  mustParse(t, "2006-01-02T15:04:05Z"),
			window: TimeWindowTimeRange{
				Begin: "3:00 PM",
				End:   "4:00 PM",
			},
			expected:      true,
			expectedError: false,
		},
		{
			name: "is within the first day of a window that spans on two days",
			now:  mustParse(t, "2006-01-02T15:04:05Z"),
			window: TimeWindowTimeRange{
				Begin: "3:00PM",
				End:   "8:00AM",
			},
			expected:      true,
			expectedError: false,
		},
		{
			name: "is within the second day of a window that spans on two days",
			now:  mustParse(t, "2006-01-02T05:04:05Z"),
			window: TimeWindowTimeRange{
				Begin: "3:00PM",
				End:   "8:00AM",
			},
			expected:      true,
			expectedError: false,
		},
		{
			name: "is outside of a window that spans on two days",
			now:  mustParse(t, "2006-01-02T10:04:05Z"),
			window: TimeWindowTimeRange{
				Begin: "3:00PM",
				End:   "8:00AM",
			},
			expected:      false,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.window.InWindow(tc.now)
			if err != nil && !tc.expectedError {
				assert.FailNow(t, err.Error())
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestInWindows(t *testing.T) {
	testCases := []struct {
		name          string
		now           time.Time
		windows       TimeWindowWhen
		expected      bool
		expectedError bool
	}{
		{
			name: "no time windows",
			now:  mustParse(t, "2006-01-02T15:04:05Z"),
			windows: TimeWindowWhen{
				Days: TimeWindowDays{},
			},
			expected:      false,
			expectedError: false,
		},
		{
			name: "is within the time window of all days",
			now:  mustParse(t, "2006-01-02T15:04:05Z"),
			windows: TimeWindowWhen{
				Days: TimeWindowDays{
					All: []*TimeWindowTimeRange{
						&TimeWindowTimeRange{
							Begin: "3:00PM",
							End:   "4:00PM",
						},
					},
				},
			},
			expected:      true,
			expectedError: false,
		},
		{
			name: "is within one of the time windows of all days",
			now:  mustParse(t, "2006-01-02T00:00:00Z"),
			windows: TimeWindowWhen{
				Days: TimeWindowDays{
					All: []*TimeWindowTimeRange{
						&TimeWindowTimeRange{
							Begin: "10:00AM",
							End:   "11:00AM",
						},
						&TimeWindowTimeRange{
							Begin: "11:00 PM",
							End:   "1:00 AM",
						},
					},
				},
			},
			expected:      true,
			expectedError: false,
		},
		{
			name: "is within one the time window of Monday",
			now:  mustParse(t, "2006-01-02T17:04:05Z"), // Weekday().String() == Monday
			windows: TimeWindowWhen{
				Days: TimeWindowDays{
					All: []*TimeWindowTimeRange{
						&TimeWindowTimeRange{
							Begin: "3:00PM",
							End:   "4:00PM",
						},
					},
					Monday: []*TimeWindowTimeRange{
						&TimeWindowTimeRange{
							Begin: "1:00PM",
							End:   "2:00PM",
						},
						&TimeWindowTimeRange{
							Begin: "5:00PM",
							End:   "6:00PM",
						},
					},
				},
			},
			expected:      true,
			expectedError: false,
		},
		{
			name: "is outside the time windows of All",
			now:  mustParse(t, "2006-01-02T17:04:05Z"),
			windows: TimeWindowWhen{
				Days: TimeWindowDays{
					All: []*TimeWindowTimeRange{
						&TimeWindowTimeRange{
							Begin: "3:00PM",
							End:   "4:00PM",
						},
					},
				},
			},
			expected:      false,
			expectedError: false,
		},
		{
			name: "is outside the time window of any days",
			now:  mustParse(t, "2006-01-02T17:04:05Z"), // .Weekday().String() == Monday
			windows: TimeWindowWhen{
				Days: TimeWindowDays{
					All: []*TimeWindowTimeRange{
						&TimeWindowTimeRange{
							Begin: "3:00PM",
							End:   "4:00PM",
						},
					},
					Tuesday: []*TimeWindowTimeRange{
						&TimeWindowTimeRange{
							Begin: "5:00PM",
							End:   "6:00PM",
						},
					},
				},
			},
			expected:      false,
			expectedError: false,
		},
		{
			name: "invalid time format",
			now:  mustParse(t, "2006-01-02T17:04:05Z"),
			windows: TimeWindowWhen{
				Days: TimeWindowDays{
					All: []*TimeWindowTimeRange{
						&TimeWindowTimeRange{
							Begin: "15:00",
							End:   "16:00",
						},
					},
				},
			},
			expected:      false,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.windows.InWindows(tc.now)
			if err != nil && !tc.expectedError {
				assert.FailNow(t, err.Error())
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}
