package graphql

import (
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityFilters(t *testing.T) {
	fs := EntityFilters()
	require.NotEmpty(t, fs)

	testCases := []struct {
		statement   string
		setupRecord func() *v2.Entity
		expect      bool
	}{
		{
			statement: "class:agent",
			expect:    true,
			setupRecord: func() *v2.Entity {
				entity := v2.FixtureEntity("a")
				entity.EntityClass = "agent"
				return entity
			},
		},
		{
			statement: "class:proxy",
			expect:    true,
			setupRecord: func() *v2.Entity {
				entity := v2.FixtureEntity("a")
				entity.EntityClass = "proxy"
				return entity
			},
		},
		{
			statement: "subscription:unix",
			expect:    true,
			setupRecord: func() *v2.Entity {
				entity := v2.FixtureEntity("a")
				entity.Subscriptions = []string{"unix", "psql"}
				return entity
			},
		},
		{
			statement: "subscription:unix",
			expect:    false,
			setupRecord: func() *v2.Entity {
				entity := v2.FixtureEntity("a")
				entity.Subscriptions = []string{"windows"}
				return entity
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
