package graphql

import (
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckTypeHistoryFieldImpl(t *testing.T) {
	testCases := []struct {
		expectedLen int
		lastArg     int
	}{
		{
			expectedLen: 21,
			lastArg:     50,
		},
		{
			expectedLen: 10,
			lastArg:     10,
		},
		{
			expectedLen: 0,
			lastArg:     0,
		},
		{
			expectedLen: 0,
			lastArg:     -10,
		},
	}

	check := types.FixtureCheck("test")
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("w/ argument of %d", tc.expectedLen), func(t *testing.T) {
			params := schema.CheckHistoryFieldResolverParams{}
			params.Source = check
			params.Args.Last = tc.lastArg

			impl := checkImpl{}
			res, err := impl.History(params)
			require.NoError(t, err)
			assert.Len(t, res, tc.expectedLen)
		})
	}
}
