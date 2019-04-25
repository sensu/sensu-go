package graphql

import (
	"context"
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/backend/apid/graphql/mockclient"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestEventTypeIsSilencedField(t *testing.T) {
	event := types.FixtureEvent("my-entity", "my-check")
	event.Check.Subscriptions = []string{"unix"}

	client, _ := client.NewClientFactory()
	client.On("ListSilenceds", mock.Anything, "", "", mock.Anything).Return([]types.Silenced{
		*types.FixtureSilenced("*:my-check"),
		*types.FixtureSilenced("unix:not-my-check"),
		*types.FixtureSilenced("entity:my-entity:*"),
	}, nil).Once()

	impl := &eventImpl{}
	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = event

	// return associated silence
	res, err := impl.IsSilenced(params)
	require.NoError(t, err)
	assert.True(t, res)
}

func TestEventTypeSilencesField(t *testing.T) {
	event := types.FixtureEvent("my-entity", "my-check")
	event.Check.Subscriptions = []string{"unix"}
	event.Entity.Subscriptions = []string{"unix"}

	client, _ := client.NewClientFactory()
	client.On("ListSilenceds", mock.Anything, "", "", mock.Anything).Return([]types.Silenced{
		*types.FixtureSilenced("*:my-check"),                    // match
		*types.FixtureSilenced("unix:my-check"),                 // match
		*types.FixtureSilenced("unix:not-my-check"),             // not match
		*types.FixtureSilenced("entity:my-entity:*"),            // match
		*types.FixtureSilenced("entity:my-entity:my-check"),     // match
		*types.FixtureSilenced("entity:my-entity:not-my-check"), // not match
		*types.FixtureSilenced("not-my-subscription:*"),         // not match
	}, nil).Once()

	impl := &eventImpl{}
	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = event

	// return associated silence
	res, err := impl.Silences(params)
	require.NoError(t, err)
	assert.Len(t, res, 4)
}
