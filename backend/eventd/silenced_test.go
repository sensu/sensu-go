package eventd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSilenced(t *testing.T) {
	testCases := []struct {
		name                  string
		event                 *types.Event
		silencedSubscriptions []*types.Silenced
		silencedChecks        []*types.Silenced
		expectedEntries       []string
	}{
		{
			name:  "Sets the silenced attribute of an event",
			event: types.FixtureEvent("foo", "check_cpu"),
			silencedSubscriptions: []*types.Silenced{},
			silencedChecks: []*types.Silenced{
				types.FixtureSilenced("entity:foo:check_cpu"),
			},
			expectedEntries: []string{"entity:foo:check_cpu"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), types.OrganizationKey, "default")
			ctx = context.WithValue(ctx, types.EnvironmentKey, "default")

			mockStore := &mockstore.MockStore{}
			mockStore.On(
				"GetSilencedEntriesBySubscription",
				mock.Anything,
			).Return(tc.silencedSubscriptions, nil)

			mockStore.On(
				"GetSilencedEntriesByCheckName",
				mock.Anything,
			).Return(tc.silencedChecks, nil)

			result := getSilenced(ctx, tc.event, mockStore)
			assert.Nil(t, result)
			assert.Equal(t, tc.expectedEntries, tc.event.Silenced)
		})
	}
}

func TestSilencedBy(t *testing.T) {
	testCases := []struct {
		name            string
		event           *types.Event
		entries         []*types.Silenced
		expectedEntries []string
	}{
		{
			name:            "no entries",
			event:           types.FixtureEvent("foo", "check_cpu"),
			expectedEntries: []string{},
		},
		{
			name:  "not silenced",
			event: types.FixtureEvent("foo", "check_cpu"),
			entries: []*types.Silenced{
				types.FixtureSilenced("entity:foo:check_mem"),
				types.FixtureSilenced("entity:bar:*"),
				types.FixtureSilenced("foo:check_cpu"),
				types.FixtureSilenced("foo:*"),
				types.FixtureSilenced("*:check_mem"),
			},
			expectedEntries: []string{},
		},
		{
			name:  "silenced by check",
			event: types.FixtureEvent("foo", "check_cpu"),
			entries: []*types.Silenced{
				types.FixtureSilenced("*:check_cpu"),
			},
			expectedEntries: []string{"*:check_cpu"},
		},
		{
			name:  "silenced by entity subscription",
			event: types.FixtureEvent("foo", "check_cpu"),
			entries: []*types.Silenced{
				types.FixtureSilenced("entity:foo:*"),
			},
			expectedEntries: []string{"entity:foo:*"},
		},
		{
			name:  "silenced by entity's check subscription",
			event: types.FixtureEvent("foo", "check_cpu"),
			entries: []*types.Silenced{
				types.FixtureSilenced("entity:foo:check_cpu"),
			},
			expectedEntries: []string{"entity:foo:check_cpu"},
		},
		{
			name:  "silenced by check subscription",
			event: types.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*types.Silenced{
				types.FixtureSilenced("linux:*"),
			},
			expectedEntries: []string{"linux:*"},
		},
		{
			name:  "silenced by subscription with check",
			event: types.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*types.Silenced{
				types.FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []string{"linux:check_cpu"},
		},
		{
			name:  "silenced by multiple entries",
			event: types.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*types.Silenced{
				types.FixtureSilenced("entity:foo:*"),
				types.FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []string{"entity:foo:*", "linux:check_cpu"},
		},
		{
			name:  "silenced by duplicated entries",
			event: types.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*types.Silenced{
				types.FixtureSilenced("entity:foo:*"),
				types.FixtureSilenced("entity:foo:*"),
			},
			expectedEntries: []string{"entity:foo:*"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := silencedBy(tc.event, tc.entries)
			assert.Equal(t, tc.expectedEntries, result)
		})
	}
}
