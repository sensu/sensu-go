package time

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestInWindow(t *testing.T) {
	testCases := []struct {
		name          string
		now           string // time.RFC3339 format, supports time.Now()
		begin         string // time.Kitchen format
		end           string // time.Kitchen format
		expected      bool
		expectedError bool
	}{
		{
			name:          "is within window",
			now:           "2006-01-02T15:04:05Z",
			begin:         "3:00PM",
			end:           "4:00PM",
			expected:      true,
			expectedError: false,
		},
		{
			name:          "is outside window",
			now:           "2006-01-02T10:04:05Z",
			begin:         "3:00PM",
			end:           "4:00PM",
			expected:      false,
			expectedError: false,
		},
		{
			name:          "unsupported time window beginning format",
			now:           "2006-01-02T10:04:05Z",
			begin:         "15:00",
			end:           "4:00PM",
			expected:      false,
			expectedError: true,
		},
		{
			name:          "unsupported time window ending format",
			now:           "2006-01-02T10:04:05Z",
			begin:         "3:00PM",
			end:           "16:00",
			expected:      false,
			expectedError: true,
		},
		{
			name:          "supports time window with whitespaces for backward compatibility",
			now:           "2006-01-02T15:04:05Z",
			begin:         "3:00 PM",
			end:           "4:00 PM",
			expected:      true,
			expectedError: false,
		},
		{
			name:          "is within the first day of a window that spans on two days",
			now:           "2006-01-02T15:04:05Z",
			begin:         "3:00PM",
			end:           "8:00AM",
			expected:      true,
			expectedError: false,
		},
		{
			name:          "is within the second day of a window that spans on two days",
			now:           "2006-01-02T05:04:05Z",
			begin:         "3:00PM",
			end:           "8:00AM",
			expected:      true,
			expectedError: false,
		},
		{
			name:          "is outside of a window that spans on two days",
			now:           "2006-01-02T10:04:05Z",
			begin:         "3:00PM",
			end:           "8:00AM",
			expected:      false,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			now, err := time.Parse(time.RFC3339, tc.now)
			if err != nil {
				assert.FailNow(t, err.Error())
			}

			result, err := InWindow(now, tc.begin, tc.end)
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
		now           string // time.RFC3339 format, supports time.Now()
		windows       types.TimeWindowWhen
		expected      bool
		expectedError bool
	}{
		{
			name: "no time windows",
			now:  "2006-01-02T15:04:05Z",
			windows: types.TimeWindowWhen{
				Days: types.TimeWindowDays{},
			},
			expected:      false,
			expectedError: false,
		},
		{
			name: "is within the time window of all days",
			now:  "2006-01-02T15:04:05Z",
			windows: types.TimeWindowWhen{
				Days: types.TimeWindowDays{
					All: []*types.TimeWindowTimeRange{
						&types.TimeWindowTimeRange{
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
			now:  "2006-01-02T15:04:05Z",
			windows: types.TimeWindowWhen{
				Days: types.TimeWindowDays{
					All: []*types.TimeWindowTimeRange{
						&types.TimeWindowTimeRange{
							Begin: "10:00AM",
							End:   "11:00AM",
						},
						&types.TimeWindowTimeRange{
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
			name: "is within one the time window of Monday",
			now:  "2006-01-02T17:04:05Z", // Weekday().String() == Monday
			windows: types.TimeWindowWhen{
				Days: types.TimeWindowDays{
					All: []*types.TimeWindowTimeRange{
						&types.TimeWindowTimeRange{
							Begin: "3:00PM",
							End:   "4:00PM",
						},
					},
					Monday: []*types.TimeWindowTimeRange{
						&types.TimeWindowTimeRange{
							Begin: "1:00PM",
							End:   "2:00PM",
						},
						&types.TimeWindowTimeRange{
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
			now:  "2006-01-02T17:04:05Z",
			windows: types.TimeWindowWhen{
				Days: types.TimeWindowDays{
					All: []*types.TimeWindowTimeRange{
						&types.TimeWindowTimeRange{
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
			now:  "2006-01-02T17:04:05Z", // .Weekday().String() == Monday
			windows: types.TimeWindowWhen{
				Days: types.TimeWindowDays{
					All: []*types.TimeWindowTimeRange{
						&types.TimeWindowTimeRange{
							Begin: "3:00PM",
							End:   "4:00PM",
						},
					},
					Tuesday: []*types.TimeWindowTimeRange{
						&types.TimeWindowTimeRange{
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
			now:  "2006-01-02T17:04:05Z",
			windows: types.TimeWindowWhen{
				Days: types.TimeWindowDays{
					All: []*types.TimeWindowTimeRange{
						&types.TimeWindowTimeRange{
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
			now, err := time.Parse(time.RFC3339, tc.now)
			if err != nil {
				assert.FailNow(t, err.Error())
			}

			result, err := InWindows(now, tc.windows)
			if err != nil && !tc.expectedError {
				assert.FailNow(t, err.Error())
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}
