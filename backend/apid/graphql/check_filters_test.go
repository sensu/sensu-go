package graphql

import (
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckFilters(t *testing.T) {
	fs := CheckFilters()
	require.NotEmpty(t, fs)

	testCases := []struct {
		statement   string
		setupRecord func() *v2.CheckConfig
		expect      bool
	}{
		{
			statement: "published:true",
			expect:    true,
			setupRecord: func() *v2.CheckConfig {
				return v2.FixtureCheckConfig("a")
			},
		},
		{
			statement: "published:False",
			expect:    false,
			setupRecord: func() *v2.CheckConfig {
				return v2.FixtureCheckConfig("a")
			},
		},
		{
			statement: "published:True",
			expect:    false,
			setupRecord: func() *v2.CheckConfig {
				chk := v2.FixtureCheckConfig("a")
				chk.Publish = false
				return chk
			},
		},
		{
			statement: "subscription:unix",
			expect:    true,
			setupRecord: func() *v2.CheckConfig {
				chk := v2.FixtureCheckConfig("a")
				chk.Subscriptions = []string{"unix", "psql"}
				return chk
			},
		},
		{
			statement: "subscription:windows",
			expect:    false,
			setupRecord: func() *v2.CheckConfig {
				chk := v2.FixtureCheckConfig("a")
				chk.Subscriptions = []string{"unix", "psql"}
				return chk
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
