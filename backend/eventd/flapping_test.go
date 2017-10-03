package eventd

import (
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestTotalStateChange(t *testing.T) {
	testCases := []struct {
		desc                     string
		event                    *types.Event
		expectedTotalStateChange int
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
					History: []types.CheckHistory{
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
					},
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
