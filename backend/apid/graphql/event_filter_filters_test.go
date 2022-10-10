package graphql

import (
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventFilterFilters(t *testing.T) {
	fs := EventFilterFilters()
	require.NotEmpty(t, fs)

	testCases := []struct {
		statement   string
		setupRecord func() *v2.EventFilter
		expect      bool
	}{
		{
			statement: "action:allow",
			expect:    true,
			setupRecord: func() *v2.EventFilter {
				filter := v2.FixtureEventFilter("a")
				filter.Action = "allow"
				return filter
			},
		},
		{
			statement: "action:deny",
			expect:    true,
			setupRecord: func() *v2.EventFilter {
				filter := v2.FixtureEventFilter("a")
				filter.Action = "deny"
				return filter
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
