package graphql

import (
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSilenceFilters(t *testing.T) {
	fs := SilenceFilters()
	require.NotEmpty(t, fs)

	testCases := []struct {
		statement   string
		setupRecord func() *v2.Silenced
		expect      bool
	}{
		{
			statement: "check:disk-check",
			expect:    true,
			setupRecord: func() *v2.Silenced {
				return v2.FixtureSilenced("*:disk-check")
			},
		},
		{
			statement: "check:test",
			expect:    false,
			setupRecord: func() *v2.Silenced {
				return v2.FixtureSilenced("other-check:*")
			},
		},
		{
			statement: "subscription:unix",
			expect:    true,
			setupRecord: func() *v2.Silenced {
				return v2.FixtureSilenced("unix:*")
			},
		},
		{
			statement: "subscription:windows",
			expect:    false,
			setupRecord: func() *v2.Silenced {
				return v2.FixtureSilenced("unix:*")
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
