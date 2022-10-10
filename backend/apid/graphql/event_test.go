package graphql

import (
	"context"
	"fmt"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEventTypeIsNewIncidentFieldImpl(t *testing.T) {
	mkEvent := func(status, lastStatus uint32) *corev2.Event {
		event := corev2.Event{
			Check: &corev2.Check{
				Status: status,
				History: []corev2.CheckHistory{
					{Status: lastStatus},
				},
			},
		}
		return &event
	}

	testCases := []struct {
		assertion      string
		expectedResult bool
		event          *corev2.Event
	}{
		{"no check", false, &corev2.Event{Check: nil, Timestamp: 12345678}},
		{"no history", false, &corev2.Event{Check: &corev2.Check{}}},
		{"currently OK", false, mkEvent(0, 0)},
		{"ongoing incident", false, mkEvent(2, 2)},
		{"ongoing incident w/ new code", false, mkEvent(2, 1)},
		{"new incident", true, mkEvent(2, 0)},
		{"new unknown incident", true, mkEvent(127, 0)},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("event %s", tc.assertion), func(t *testing.T) {
			params := graphql.ResolveParams{Context: context.Background()}
			params.Source = tc.event

			impl := eventImpl{}
			res, err := impl.IsNewIncident(params)
			require.NoError(t, err)
			assert.Equal(t, res, tc.expectedResult)
		})
	}
}

func TestEventTypeIsSilencedField(t *testing.T) {
	event := corev2.FixtureEvent("my-entity", "my-check")
	event.Check.Subscriptions = []string{"unix"}

	client := new(MockSilencedClient)
	client.On("ListSilenced", mock.Anything).Return([]*corev2.Silenced{
		corev2.FixtureSilenced("*:my-check"),
		corev2.FixtureSilenced("unix:not-my-check"),
		corev2.FixtureSilenced("entity:my-entity:*"),
	}, nil).Once()

	impl := &eventImpl{}
	params := graphql.ResolveParams{}
	cfg := ServiceConfig{SilencedClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = event

	// return associated silence
	res, err := impl.IsSilenced(params)
	require.NoError(t, err)
	assert.True(t, res)
}

func TestEventTypeSilencesField(t *testing.T) {
	event := corev2.FixtureEvent("my-entity", "my-check")
	event.Check.Subscriptions = []string{"unix"}
	event.Entity.Subscriptions = []string{"unix"}

	client := new(MockSilencedClient)
	client.On("ListSilenced", mock.Anything).Return([]*corev2.Silenced{
		corev2.FixtureSilenced("*:my-check"),                    // match
		corev2.FixtureSilenced("unix:my-check"),                 // match
		corev2.FixtureSilenced("unix:not-my-check"),             // not match
		corev2.FixtureSilenced("entity:my-entity:*"),            // match
		corev2.FixtureSilenced("entity:my-entity:my-check"),     // match
		corev2.FixtureSilenced("entity:my-entity:not-my-check"), // not match
		corev2.FixtureSilenced("not-my-subscription:*"),         // not match
	}, nil).Once()

	impl := &eventImpl{}
	params := graphql.ResolveParams{}
	cfg := ServiceConfig{SilencedClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = event

	// return associated silence
	res, err := impl.Silences(params)
	require.NoError(t, err)
	assert.Len(t, res, 4)
}
