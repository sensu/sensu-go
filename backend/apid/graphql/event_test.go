package graphql

import (
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventTypeIsNewIncidentFieldImpl(t *testing.T) {
	mkEvent := func(status, lastStatus uint32) *types.Event {
		event := types.Event{
			Check: &types.Check{
				Status: status,
				History: []types.CheckHistory{
					{Status: lastStatus},
				},
			},
		}
		return &event
	}

	testCases := []struct {
		assertion      string
		expectedResult bool
		event          *types.Event
	}{
		{"no check", false, &types.Event{Check: nil, Timestamp: 12345678}},
		{"no history", false, &types.Event{Check: &types.Check{}}},
		{"currently OK", false, mkEvent(0, 0)},
		{"ongoing incident", false, mkEvent(2, 2)},
		{"ongoing incident w/ new code", false, mkEvent(2, 1)},
		{"new incident", true, mkEvent(2, 0)},
		{"new unknown incident", true, mkEvent(127, 0)},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("event %s", tc.assertion), func(t *testing.T) {
			params := graphql.ResolveParams{}
			params.Source = tc.event

			impl := eventImpl{}
			res, err := impl.IsNewIncident(params)
			require.NoError(t, err)
			assert.Equal(t, res, tc.expectedResult)
		})
	}
}
