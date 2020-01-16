package v2

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func fictionalHistory() []CheckHistory {
	return []CheckHistory{
		CheckHistory{Status: 0},
		CheckHistory{Status: 0},
		CheckHistory{Status: 0},
		CheckHistory{Status: 1},
		CheckHistory{Status: 2},
		CheckHistory{Status: 3},
		CheckHistory{Status: 3},
		CheckHistory{Status: 3},
		CheckHistory{Status: 3},
		CheckHistory{Status: 0},
		CheckHistory{Status: 0},
		CheckHistory{Status: 0},
		CheckHistory{Status: 1},
		CheckHistory{Status: 1},
		CheckHistory{Status: 1},
		CheckHistory{Status: 1},
		CheckHistory{Status: 0},
		CheckHistory{Status: 0},
		CheckHistory{Status: 0},
		CheckHistory{Status: 2},
		CheckHistory{Status: 2},
	}
}

func TestIsFlapping(t *testing.T) {
	testCases := []struct {
		desc             string
		event            *Event
		expectedFlapping bool
	}{
		{
			"low_flap_threshold not configured",
			&Event{
				Check: &Check{},
			},
			false,
		},
		{
			"high_flap_threshold not configured",
			&Event{
				Check: &Check{
					LowFlapThreshold: 10,
				},
			},
			false,
		},
		{
			"check is still flapping",
			&Event{
				Check: &Check{
					LowFlapThreshold:  10,
					HighFlapThreshold: 30,
					State:             EventFlappingState,
					TotalStateChange:  15,
					History: []CheckHistory{
						CheckHistory{Flapping: true},
						CheckHistory{Flapping: true},
					},
				},
			},
			true,
		},
		{
			"check is no longer flapping",
			&Event{
				Check: &Check{
					LowFlapThreshold:  10,
					HighFlapThreshold: 30,
					State:             EventFlappingState,
					History: []CheckHistory{
						CheckHistory{Flapping: true},
						CheckHistory{Flapping: true},
					},
					TotalStateChange: 5,
				},
			},
			false,
		},
		{
			"check is now flapping",
			&Event{
				Check: &Check{
					LowFlapThreshold:  10,
					HighFlapThreshold: 30,
					State:             EventFailingState,
					TotalStateChange:  35,
					History: []CheckHistory{
						CheckHistory{Flapping: false},
						CheckHistory{Flapping: false},
					},
				},
			},
			true,
		},
		{
			"check is not flapping",
			&Event{
				Check: &Check{
					State:             EventPassingState,
					LowFlapThreshold:  10,
					HighFlapThreshold: 30,
					TotalStateChange:  5,
					History: []CheckHistory{
						CheckHistory{Flapping: false},
						CheckHistory{Flapping: false},
					},
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := isFlapping(tc.event.Check)
			assert.Equal(t, tc.expectedFlapping, result)
		})
	}
}

func TestTotalStateChange(t *testing.T) {
	testCases := []struct {
		desc                     string
		event                    *Event
		expectedTotalStateChange uint32
	}{
		{
			"with less than 21 check result",
			&Event{
				Check: &Check{
					History: make([]CheckHistory, 20),
				},
			},
			0,
		},
		{
			"with no changes",
			&Event{
				Check: &Check{
					History: []CheckHistory{
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
						CheckHistory{Status: 0},
					},
				},
			},
			0,
		},
		{
			"and weights the last 21 check result",
			&Event{
				Check: &Check{
					History: fictionalHistory(),
				},
			},
			34,
		},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf("calculate total state change %s", tc.desc)
		t.Run(testName, func(t *testing.T) {
			result := totalStateChange(tc.event.Check)
			assert.Equal(t, tc.expectedTotalStateChange, result)
		})
	}
}
