package graphql

import (
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerFilters(t *testing.T) {
	fs := HandlerFilters()
	require.NotEmpty(t, fs)

	testCases := []struct {
		statement   string
		setupRecord func() *v2.Handler
		expect      bool
	}{
		{
			statement: "type:pipe",
			expect:    true,
			setupRecord: func() *v2.Handler {
				hdr := v2.FixtureHandler("a")
				hdr.Type = "pipe"
				return hdr
			},
		},
		{
			statement: "type:socket",
			expect:    true,
			setupRecord: func() *v2.Handler {
				hdr := v2.FixtureHandler("a")
				hdr.Type = "socket"
				return hdr
			},
		},
		{
			statement: "type:set",
			expect:    true,
			setupRecord: func() *v2.Handler {
				hdr := v2.FixtureHandler("a")
				hdr.Type = "set"
				return hdr
			},
		},
		{
			statement: "type:notset",
			expect:    false,
			setupRecord: func() *v2.Handler {
				hdr := v2.FixtureHandler("a")
				hdr.Type = "pipe"
				return hdr
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
