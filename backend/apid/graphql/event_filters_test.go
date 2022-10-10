package graphql

import (
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventFilters(t *testing.T) {
	fs := EventFilters()
	require.NotEmpty(t, fs)

	testCases := []struct {
		statement   string
		setupRecord func() *v2.Event
		expect      bool
	}{
		{
			statement: "status:passing",
			expect:    true,
			setupRecord: func() *v2.Event {
				return v2.FixtureEvent("a", "b")
			},
		},
		{
			statement: "status:warning",
			expect:    false,
			setupRecord: func() *v2.Event {
				return v2.FixtureEvent("a", "b")
			},
		},
		{
			statement: "status:critical",
			expect:    false,
			setupRecord: func() *v2.Event {
				return v2.FixtureEvent("a", "b")
			},
		},
		{
			statement: "status:unknown",
			expect:    false,
			setupRecord: func() *v2.Event {
				return v2.FixtureEvent("a", "b")
			},
		},
		{
			statement: "status:incident",
			expect:    false,
			setupRecord: func() *v2.Event {
				return v2.FixtureEvent("a", "b")
			},
		},
		{
			statement: "check:my-check",
			expect:    true,
			setupRecord: func() *v2.Event {
				return v2.FixtureEvent("my-entity", "my-check")
			},
		},
		{
			statement: "entity:my-entity",
			expect:    true,
			setupRecord: func() *v2.Event {
				return v2.FixtureEvent("my-entity", "my-check")
			},
		},
		{
			statement: "silenced:False",
			expect:    true,
			setupRecord: func() *v2.Event {
				return v2.FixtureEvent("a", "b")
			},
		},
		{
			statement: "silenced:true",
			expect:    true,
			setupRecord: func() *v2.Event {
				ev := v2.FixtureEvent("a", "b")
				ev.Check.Silenced = []string{"x"}
				return ev
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.statement, func(t *testing.T) {
			matches, err := filter.Compile([]string{tc.statement}, fs, nil)
			require.NoError(t, err)

			record := tc.setupRecord()
			assert.Equal(t, tc.expect, matches(record))
		})
	}
}
