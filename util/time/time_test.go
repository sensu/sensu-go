package time

import (
	"testing"
	"time"

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
