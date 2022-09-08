package v2

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
						{
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
						{
							Begin: "10:00AM",
							End:   "11:00AM",
						}, {
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
						{
							Begin: "3:00PM",
							End:   "4:00PM",
						},
					},
					Monday: []*TimeWindowTimeRange{
						{
							Begin: "1:00PM",
							End:   "2:00PM",
						}, {
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
						{
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
						{
							Begin: "3:00PM",
							End:   "4:00PM",
						},
					},
					Tuesday: []*TimeWindowTimeRange{
						{
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
						{
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

func TestTimeWindowRepeated_InWindows(t *testing.T) {
	currentTime, err := time.Parse(time.RFC3339, "2022-03-28T15:09:08+06:00")
	assert.NoError(t, err)
	assert.NotNil(t, currentTime)
	fmt.Println("current", currentTime, currentTime.Weekday())

	now := time.Now().In(currentTime.Location())
	fmt.Println("now", now, now.Weekday())
	location := currentTime.Location()
	fmt.Println("location", location.String())
	assert.NotNil(t, location)
}

func parseTime(t *testing.T, s string) time.Time {
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

func TestTimeWindowRepeated_InDayTimeRange(t *testing.T) {

	tests := []struct {
		name           string
		beginTime      string
		endTime        string
		actualTime     time.Time
		weekday        time.Weekday
		expectedResult bool
	}{
		{
			"simple positive",
			"2022-03-22T09:00:00+00:00", // tuesday
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T10:00:00+00:00"),
			time.Tuesday,
			true,
		}, {
			"simple negative lower hour",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T08:00:00+00:00"),
			time.Tuesday,
			false,
		}, {
			"simple negative higher hour",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T13:00:00+00:00"),
			time.Tuesday,
			false,
		}, {
			"simple negative lower minutes",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T08:59:00+00:00"),
			time.Tuesday,
			false,
		}, {
			"simple negative higher minutes",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T11:01:00+00:00"),
			time.Tuesday,
			false,
		}, {
			"simple negative lower seconds",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T08:59:59+00:00"),
			time.Tuesday,
			false,
		}, {
			"simple negative higher seconds",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T11:00:01+00:00"),
			time.Tuesday,
			false,
		}, {
			"simple negative different day before",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-21T10:00:00+00:00"), // monday
			time.Tuesday,
			false,
		}, {
			"simple negative different day after",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-23T10:00:00+00:00"), // wednesday
			time.Tuesday,
			false,
		}, {
			"positive next week",
			"2022-03-22T09:00:00-07:00",
			"2022-03-22T11:00:00-07:00",
			parseTime(t, "2022-03-29T10:00:00-07:00"), // wednesday
			time.Tuesday,
			true,
		}, {
			"negative next week before",
			"2022-03-22T09:00:00-07:00",
			"2022-03-22T11:00:00-07:00",
			parseTime(t, "2022-03-27T10:00:00-07:00"), // wednesday
			time.Tuesday,
			false,
		}, {
			"negative next week after",
			"2022-03-22T09:00:00-07:00",
			"2022-03-22T11:00:00-07:00",
			parseTime(t, "2022-03-30T10:00:00-07:00"), // wednesday
			time.Tuesday,
			false,
		}, {
			"positive next week different time offset",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-29T06:00:00-04:00"), // wednesday
			time.Tuesday,
			true,
		}, {
			"positive multi days",
			"2022-03-22T09:00:00+00:00",
			"2022-03-24T11:00:00+00:00",
			parseTime(t, "2022-03-30T06:00:00-04:00"), // wednesday
			time.Tuesday,
			true,
		}, {
			"negative previous week",
			"2022-03-22T09:00:00-04:00",
			"2022-03-22T11:00:00-04:00",
			parseTime(t, "2022-03-15T10:00:00-04:00"), // tuesday
			time.Tuesday,
			false,
		}, {
			"GH-4847 lower bound day time range",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T09:00:00+00:00"),
			time.Tuesday,
			true,
		}, {
			"GH-4847 upper bound day time range",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T11:00:01+00:00").Add(-1 * time.Nanosecond),
			time.Tuesday,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			window := TimeWindowRepeated{
				Begin: test.beginTime,
				End:   test.endTime,
			}

			assert.Equal(t, test.expectedResult, window.inDayTimeRange(test.actualTime, test.weekday))
		})
	}
}

func TestTimeWindowRepeated_InTimeRange(t *testing.T) {
	tests := []struct {
		name           string
		beginTime      string
		endTime        string
		actualTime     time.Time
		expectedResult bool
	}{
		{
			"simple positive",
			"2022-03-22T09:00:00+00:00", // tuesday
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T10:00:00+00:00"),
			true,
		}, {
			"negative before valid times",
			"2022-03-22T09:00:00-07:00", // tuesday
			"2022-03-22T11:00:00-07:00",
			parseTime(t, "2022-01-12T10:00:00-07:00"),
			false,
		}, {
			"simple negative before",
			"2022-03-22T09:00:00+00:00", // tuesday
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T08:00:00+00:00"),
			false,
		}, {
			"simple negative after",
			"2022-03-22T09:00:00+00:00", // tuesday
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T11:05:00+00:00"),
			false,
		}, {
			"positive after different day",
			"2022-03-22T09:00:00-07:00", // tuesday
			"2022-03-22T11:00:00-07:00",
			parseTime(t, "2023-01-12T10:00:00-07:00"),
			true,
		}, {
			"negative before different day",
			"2022-03-22T09:00:00-07:00", // tuesday
			"2022-03-22T11:00:00-07:00",
			parseTime(t, "2021-01-12T12:00:00-07:00"),
			false,
		}, {
			"negative after different day",
			"2022-03-22T09:00:00-07:00", // tuesday
			"2022-03-22T11:00:00-07:00",
			parseTime(t, "2023-01-12T08:59:59-07:00"),
			false,
		}, {
			"GH-4847 lower bound time range",
			"2022-03-22T09:00:00+00:00", // tuesday
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T09:00:00+00:00"),
			true,
		}, {
			"GH-4847 upper bound time range",
			"2022-03-22T09:00:00+00:00", // tuesday
			"2022-03-22T11:00:00+00:00",
			parseTime(t, "2022-03-22T11:00:01+00:00").Add(-1 * time.Nanosecond),
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			window := TimeWindowRepeated{
				Begin: test.beginTime,
				End:   test.endTime,
			}

			assert.Equal(t, test.expectedResult, window.inTimeRange(test.actualTime))
		})
	}
}

func TestTimeWindowRepeated_InMonthlyTimeRange(t *testing.T) {
	tests := []struct {
		name           string
		beginTime      string
		endTime        string
		actualTime     string
		expectedResult bool
	}{
		{
			"simple positive",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2022-03-22T10:00:00+00:00",
			true,
		}, {
			"previous month same day",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2022-02-22T10:00:00+00:00",
			false,
		}, {
			"next month same day",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2022-04-22T10:00:00+00:00",
			true,
		}, {
			"next year same day",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2023-04-22T10:00:00+00:00",
			true,
		}, {
			"same day different time",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2022-03-22T11:01:00+00:00",
			false,
		}, {
			"next month same day different time before",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2022-04-22T08:00:00+00:00",
			false,
		}, {
			"next month same day different time after",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2022-04-22T11:02:00+00:00",
			false,
		}, {
			"next month different day before",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2022-04-21T10:00:00+00:00",
			false,
		}, {
			"next month different day after",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2022-04-23T10:00:00+00:00",
			false,
		}, {
			"next year different day before",
			"2022-03-22T09:00:00+00:00", // tuesday
			"2022-03-22T11:00:00+00:00",
			"2023-04-20T10:00:00+00:00",
			false,
		}, {
			"next year different day before",
			"2022-03-22T09:00:00+00:00", // tuesday
			"2022-03-22T11:00:00+00:00",
			"2023-04-23T10:00:00+00:00",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			window := TimeWindowRepeated{
				Begin: test.beginTime,
				End:   test.endTime,
			}

			actualTime, err := time.Parse(time.RFC3339, test.actualTime)
			require.NoError(t, err)

			assert.Equal(t, test.expectedResult, window.inMonthlyTimeRange(actualTime))
		})
	}
}

func TestTimeWindowRepeated_InYearlyTimeRange(t *testing.T) {
	tests := []struct {
		name           string
		beginTime      string
		endTime        string
		actualTime     string
		expectedResult bool
	}{
		{
			"simple positive",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2022-03-22T10:00:00+00:00",
			true,
		}, {
			"positive next year",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2023-03-22T10:00:00+00:00",
			true,
		}, {
			"negative previous year matching date and time",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2021-03-22T10:00:00+00:00",
			false,
		}, {
			"this year different date before",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2023-03-21T10:00:00+00:00",
			false,
		}, {
			"this year different date after",
			"2022-03-22T09:00:00+00:00",
			"2022-03-22T11:00:00+00:00",
			"2023-03-23T10:00:00+00:00",
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			window := TimeWindowRepeated{
				Begin: test.beginTime,
				End:   test.endTime,
			}

			actualTime, err := time.Parse(time.RFC3339, test.actualTime)
			require.NoError(t, err)

			assert.Equal(t, test.expectedResult, window.inYearlyTimeRange(actualTime))
		})
	}
}
