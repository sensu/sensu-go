package silenced

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/stretchr/testify/assert"
)

func TestGetSilenced(t *testing.T) {
	testCases := []struct {
		name            string
		event           *corev2.Event
		silencedEntries []corev2.Resource
		expectedEntries []string
	}{
		{
			name:  "Sets the silenced attribute of an event",
			event: corev2.FixtureEvent("foo", "check_cpu"),
			silencedEntries: []corev2.Resource{
				corev2.FixtureSilenced("entity:foo:check_cpu"),
			},
			expectedEntries: []string{"entity:foo:check_cpu"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), corev2.NamespaceKey, "default")
			c := cache.NewFromResources(tc.silencedEntries, false)

			GetSilenced(ctx, tc.event, c)
			assert.Equal(t, tc.expectedEntries, tc.event.Check.Silenced)
		})
	}
}

func TestSilencedBy(t *testing.T) {
	testCases := []struct {
		name            string
		event           *corev2.Event
		entries         []*corev2.Silenced
		expectedEntries []string
	}{
		{
			name:            "no entries",
			event:           corev2.FixtureEvent("foo", "check_cpu"),
			expectedEntries: []string{},
		},
		{
			name:  "not silenced",
			event: corev2.FixtureEvent("foo", "check_cpu"),
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("entity:foo:check_mem"),
				corev2.FixtureSilenced("entity:bar:*"),
				corev2.FixtureSilenced("foo:check_cpu"),
				corev2.FixtureSilenced("foo:*"),
				corev2.FixtureSilenced("*:check_mem"),
			},
			expectedEntries: []string{},
		},
		{
			name:  "silenced by check",
			event: corev2.FixtureEvent("foo", "check_cpu"),
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("*:check_cpu"),
			},
			expectedEntries: []string{"*:check_cpu"},
		},
		{
			name:  "silenced by entity subscription",
			event: corev2.FixtureEvent("foo", "check_cpu"),
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("entity:foo:*"),
			},
			expectedEntries: []string{"entity:foo:*"},
		},
		{
			name:  "silenced by entity's check subscription",
			event: corev2.FixtureEvent("foo", "check_cpu"),
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("entity:foo:check_cpu"),
			},
			expectedEntries: []string{"entity:foo:check_cpu"},
		},
		{
			name:  "silenced by check subscription",
			event: corev2.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("linux:*"),
			},
			expectedEntries: []string{"linux:*"},
		},
		{
			name:  "silenced by subscription with check",
			event: corev2.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []string{"linux:check_cpu"},
		},
		{
			name:  "silenced by multiple entries",
			event: corev2.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("entity:foo:*"),
				corev2.FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []string{"entity:foo:*", "linux:check_cpu"},
		},
		{
			name:  "silenced by duplicated entries",
			event: corev2.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("entity:foo:*"),
				corev2.FixtureSilenced("entity:foo:*"),
			},
			expectedEntries: []string{"entity:foo:*"},
		},
		{
			name: "not silenced, silenced & client don't have a common subscription",
			event: &corev2.Event{
				Check: &corev2.Check{
					ObjectMeta: corev2.ObjectMeta{
						Name: "check_cpu",
					},
					Subscriptions: []string{"linux", "windows"},
				},
				Entity: &corev2.Entity{
					ObjectMeta: corev2.ObjectMeta{
						Name: "foo",
					},
					Subscriptions: []string{"linux"},
				},
			},
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("windows:check_cpu"),
			},
			expectedEntries: []string{},
		},
		{
			name: "silenced, silenced & client do have a common subscription",
			event: &corev2.Event{
				Check: &corev2.Check{
					ObjectMeta: corev2.ObjectMeta{
						Name: "check_cpu",
					},
					Subscriptions: []string{"linux", "windows"},
				},
				Entity: &corev2.Entity{
					ObjectMeta: corev2.ObjectMeta{
						Name: "foo",
					},
					Subscriptions: []string{"linux"},
				},
			},
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []string{"linux:check_cpu"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SilencedBy(tc.event, tc.entries)
			assert.Equal(t, tc.expectedEntries, result)
		})
	}
}
