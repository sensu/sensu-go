package eventd

import (
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func fictionalHistory() []types.CheckHistory {
	return []types.CheckHistory{
		types.CheckHistory{Status: 0},
		types.CheckHistory{Status: 0},
		types.CheckHistory{Status: 0},
		types.CheckHistory{Status: 1},
		types.CheckHistory{Status: 3},
		types.CheckHistory{Status: 2},
		types.CheckHistory{Status: 2},
		types.CheckHistory{Status: 2},
		types.CheckHistory{Status: 2},
		types.CheckHistory{Status: 0},
		types.CheckHistory{Status: 0},
		types.CheckHistory{Status: 0},
		types.CheckHistory{Status: 1},
		types.CheckHistory{Status: 1},
		types.CheckHistory{Status: 1},
		types.CheckHistory{Status: 1},
		types.CheckHistory{Status: 0},
		types.CheckHistory{Status: 0},
		types.CheckHistory{Status: 0},
		types.CheckHistory{Status: 2},
		types.CheckHistory{Status: 2},
	}
}

func TestIsFlapping(t *testing.T) {
	testCases := []struct {
		desc             string
		event            *types.Event
		expectedFlapping bool
	}{
		{
			"low_flap_threshold not configured",
			&types.Event{
				Check: &types.Check{},
			},
			false,
		},
		{
			"high_flap_threshold not configured",
			&types.Event{
				Check: &types.Check{
					LowFlapThreshold: 10,
				},
			},
			false,
		},
		{
			"check is still flapping",
			&types.Event{
				Check: &types.Check{
					LowFlapThreshold:  10,
					HighFlapThreshold: 30,
					State:             types.EventFlappingState,
					TotalStateChange:  15,
				},
			},
			true,
		},
		{
			"check is no longer flapping",
			&types.Event{
				Check: &types.Check{
					LowFlapThreshold:  10,
					HighFlapThreshold: 30,
					State:             types.EventFlappingState,
					TotalStateChange:  5,
				},
			},
			false,
		},
		{
			"check is now flapping",
			&types.Event{
				Check: &types.Check{
					LowFlapThreshold:  10,
					HighFlapThreshold: 30,
					State:             types.EventFailingState,
					TotalStateChange:  35,
				},
			},
			true,
		},
		{
			"check is not flapping",
			&types.Event{
				Check: &types.Check{
					State:             types.EventPassingState,
					LowFlapThreshold:  10,
					HighFlapThreshold: 30,
					TotalStateChange:  5,
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := isFlapping(tc.event)
			assert.Equal(t, tc.expectedFlapping, result)
		})
	}
}

func TestTotalStateChange(t *testing.T) {
	testCases := []struct {
		desc                     string
		event                    *types.Event
		expectedTotalStateChange uint32
	}{
		{
			"with less than 21 check result",
			&types.Event{
				Check: &types.Check{
					History: make([]types.CheckHistory, 20),
				},
			},
			0,
		},
		{
			"with no changes",
			&types.Event{
				Check: &types.Check{
					History: []types.CheckHistory{
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
						types.CheckHistory{Status: 0},
					},
				},
			},
			0,
		},
		{
			"and weights the last 21 check result",
			&types.Event{
				Check: &types.Check{
					History: fictionalHistory(),
				},
			},
			34,
		},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf("calculate total state change %s", tc.desc)
		t.Run(testName, func(t *testing.T) {
			result := totalStateChange(tc.event)
			assert.Equal(t, tc.expectedTotalStateChange, result)
		})
	}
}
