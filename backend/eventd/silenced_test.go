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
			expectedEntries: []string{"entity:foo:*"},
		},
		{
			name: "not silenced, silenced & client don't have a common subscription",
			event: &types.Event{
				Check: &types.Check{
					Name:          "check_cpu",
					Subscriptions: []string{"linux", "windows"},
				},
				Entity: &types.Entity{
					ID:            "foo",
					Subscriptions: []string{"linux"},
				},
			},
			entries: []*types.Silenced{
				types.FixtureSilenced("windows:check_cpu"),
			},
			expectedEntries: []string{},
		},
		{
			name: "silenced, silenced & client do have a common subscription",
			event: &types.Event{
				Check: &types.Check{
					Name:          "check_cpu",
					Subscriptions: []string{"linux", "windows"},
				},
				Entity: &types.Entity{
					ID:            "foo",
					Subscriptions: []string{"linux"},
				},
			},
			entries: []*types.Silenced{
				types.FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []string{"linux:check_cpu"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := silencedBy(tc.event, tc.entries)
			assert.Equal(t, tc.expectedEntries, result)
		})
	}
}

func TestHandleExpireOnResolveEntries(t *testing.T) {
	expireOnResolve := func(s *types.Silenced) *types.Silenced {
		s.ExpireOnResolve = true
		return s
	}

	resolution := func(e *types.Event) *types.Event {
		e.Check.History = []types.CheckHistory{
			types.CheckHistory{Status: 1},
		}
		e.Check.Status = 0
		return e
	}

	testCases := []struct {
		name                    string
		event                   *types.Event
		silencedEntry           *types.Silenced
		expectedSilencedEntries []string
	}{
		{
			name:                    "Non-resolution Non-expire-on-resolve Event",
			event:                   types.FixtureEvent("entity1", "check1"),
			silencedEntry:           types.FixtureSilenced("sub1:check1"),
			expectedSilencedEntries: []string{"sub1:check1"},
		},
		{
			name:                    "Non-Resolution Expire-on-resolve Event",
			event:                   types.FixtureEvent("entity1", "check1"),
			silencedEntry:           expireOnResolve(types.FixtureSilenced("sub1:check1")),
			expectedSilencedEntries: []string{"sub1:check1"},
		},
		{
			name:                    "Resolution Non-expire-on-resolve Event",
			event:                   resolution(types.FixtureEvent("entity1", "check1")),
			silencedEntry:           types.FixtureSilenced("sub1:check1"),
			expectedSilencedEntries: []string{"sub1:check1"},
		},
		{
			name:                    "Resolution Expire-on-resolve Event",
			event:                   resolution(types.FixtureEvent("entity1", "check1")),
			silencedEntry:           expireOnResolve(types.FixtureSilenced("sub1:check1")),
			expectedSilencedEntries: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), types.OrganizationKey, "default")
			ctx = context.WithValue(ctx, types.EnvironmentKey, "default")

			mockStore := &mockstore.MockStore{}

			mockStore.On(
				"GetSilencedEntryByID",
				mock.Anything,
				mock.Anything,
			).Return(tc.silencedEntry, nil)

			mockStore.On(
				"DeleteSilencedEntryByID",
				mock.Anything,
				mock.Anything,
			).Return(nil)

			tc.event.Silenced = []string{tc.silencedEntry.ID}

			err := handleExpireOnResolveEntries(ctx, tc.event, mockStore)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedSilencedEntries, tc.event.Silenced)
		})
	}
}
