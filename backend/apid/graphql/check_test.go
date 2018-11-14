package graphql

import (
	"fmt"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckTypeHistoryFieldImpl(t *testing.T) {
	testCases := []struct {
		expectedLen int
		firstArg    int
	}{
		{
			expectedLen: 21,
			firstArg:    50,
		},
		{
			expectedLen: 10,
			firstArg:    10,
		},
		{
			expectedLen: 0,
			firstArg:    0,
		},
		{
			expectedLen: 0,
			firstArg:    -10,
		},
	}

	check := types.FixtureCheck("test")
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("w/ argument of %d", tc.expectedLen), func(t *testing.T) {
			params := schema.CheckHistoryFieldResolverParams{}
			params.Source = check
			params.Args.First = tc.firstArg

			impl := checkImpl{}
			res, err := impl.History(params)
			require.NoError(t, err)
			assert.Len(t, res, tc.expectedLen)
		})
	}
}

func TestCheckTypeLastOKFieldImpl(t *testing.T) {
	now := time.Now()
	check := types.FixtureCheck("test")
	check.LastOK = now.Unix()

	impl := checkImpl{}
	params := graphql.ResolveParams{Source: check}

	res, err := impl.LastOK(params)
	require.NoError(t, err)
	assert.Equal(t, now.Unix(), res.Unix())
}

func TestCheckTypeIssuedFieldImpl(t *testing.T) {
	now := time.Now()
	check := types.FixtureCheck("test")
	check.Issued = now.Unix()

	impl := checkImpl{}
	params := graphql.ResolveParams{Source: check}

	res, err := impl.Issued(params)
	require.NoError(t, err)
	assert.Equal(t, now.Unix(), res.Unix())
}

func TestCheckTypeNodeIDFieldImpl(t *testing.T) {
	check := types.FixtureCheck("test")
	params := graphql.ResolveParams{Source: check}

	impl := checkImpl{}
	res, err := impl.NodeID(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestCheckTypeIsSilencedField(t *testing.T) {
	check := types.FixtureCheck("my-check")
	check.Silenced = []string{"unix:my-check"}
	mock := mockSilenceQuerier{els: []*types.Silenced{
		types.FixtureSilenced("unix:my-check"),
	}}

	// return associated silence
	impl := &checkImpl{silenceQuerier: mock}
	res, err := impl.IsSilenced(graphql.ResolveParams{Source: check})
	require.NoError(t, err)
	assert.True(t, res)
}

func TestCheckTypeSilencesField(t *testing.T) {
	check := types.FixtureCheck("my-check")
	check.Subscriptions = []string{"unix"}
	check.Silenced = []string{"unix:my-check"}
	mock := mockSilenceQuerier{els: []*types.Silenced{
		types.FixtureSilenced("unix:my-check"),
		types.FixtureSilenced("fred:my-check"),
		types.FixtureSilenced("unix:not-my-check"),
	}}

	// return associated silence
	impl := &checkImpl{silenceQuerier: mock}
	res, err := impl.Silences(graphql.ResolveParams{Source: check})
	require.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestCheckTypeRuntimeAssetsField(t *testing.T) {
	check := types.FixtureCheck("my-check")
	check.RuntimeAssets = []string{"one", "two"}
	mock := mockAssetQuerier{els: []*types.Asset{
		types.FixtureAsset("one"),
		types.FixtureAsset("two"),
		types.FixtureAsset("three"),
	}}

	// return associated silence
	impl := &checkImpl{assetQuerier: mock}
	res, err := impl.RuntimeAssets(graphql.ResolveParams{Source: check})
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckConfigTypeIsSilencedField(t *testing.T) {
	check := types.FixtureCheckConfig("my-check")
	check.Subscriptions = []string{"unix"}
	mock := mockSilenceQuerier{els: []*types.Silenced{
		types.FixtureSilenced("*:my-check"),
	}}

	// return associated silence
	impl := &checkCfgImpl{silenceQuerier: mock}
	res, err := impl.IsSilenced(graphql.ResolveParams{Source: check})
	require.NoError(t, err)
	assert.True(t, res)
}

func TestCheckConfigTypeSilencesField(t *testing.T) {
	check := types.FixtureCheckConfig("my-check")
	check.Subscriptions = []string{"unix"}
	mock := mockSilenceQuerier{els: []*types.Silenced{
		types.FixtureSilenced("*:my-check"),
		types.FixtureSilenced("unix:*"),
		types.FixtureSilenced("unix:my-check"),
		types.FixtureSilenced("unix:different-check"),
		types.FixtureSilenced("unrelated:my-check"),
		types.FixtureSilenced("*:another-check"),
	}}

	// return associated silence
	impl := &checkCfgImpl{silenceQuerier: mock}
	res, err := impl.Silences(graphql.ResolveParams{Source: check})
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckConfigTypeRuntimeAssetsField(t *testing.T) {
	check := types.FixtureCheckConfig("my-check")
	check.RuntimeAssets = []string{"one", "two"}
	mock := mockAssetQuerier{els: []*types.Asset{
		types.FixtureAsset("one"),
		types.FixtureAsset("two"),
		types.FixtureAsset("three"),
	}}

	// return associated silence
	impl := &checkCfgImpl{assetQuerier: mock}
	res, err := impl.RuntimeAssets(graphql.ResolveParams{Source: check})
	require.NoError(t, err)
	assert.Len(t, res, 2)
}
