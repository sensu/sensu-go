package graphql

import (
	"context"
	"fmt"
	"testing"
	"time"

	v2 "github.com/sensu/sensu-go/api/core/v2"
	client "github.com/sensu/sensu-go/backend/apid/graphql/mockclient"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	// return associated silence
	impl := &checkImpl{}
	res, err := impl.IsSilenced(graphql.ResolveParams{Source: check})
	require.NoError(t, err)
	assert.True(t, res)
}

func TestCheckTypeSilencesField(t *testing.T) {
	check := types.FixtureCheck("my-check")
	check.Subscriptions = []string{"unix"}
	check.Silenced = []string{"unix:my-check"}

	client, _ := client.NewClientFactory()
	client.On("ListSilenceds", mock.Anything, "", "", mock.Anything).Return([]types.Silenced{
		*types.FixtureSilenced("unix:my-check"),
		*types.FixtureSilenced("fred:my-check"),
		*types.FixtureSilenced("unix:not-my-check"),
	}, nil).Once()

	impl := &checkImpl{}
	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = check

	// return associated silence
	res, err := impl.Silences(params)
	require.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestCheckTypeRuntimeAssetsField(t *testing.T) {
	check := types.FixtureCheck("my-check")
	check.RuntimeAssets = []string{"one", "two"}

	client, _ := client.NewClientFactory()
	client.On("ListAssets", mock.Anything, mock.Anything).Return([]types.Asset{
		*types.FixtureAsset("one"),
		*types.FixtureAsset("two"),
		*types.FixtureAsset("three"),
	}, nil).Once()

	// return associated silence
	impl := &checkImpl{}
	ctx := contextWithLoaders(context.Background(), client)
	res, err := impl.RuntimeAssets(graphql.ResolveParams{Source: check, Context: ctx})
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckConfigTypeIsSilencedField(t *testing.T) {
	check := types.FixtureCheckConfig("my-check")
	check.Subscriptions = []string{"unix"}

	client, _ := client.NewClientFactory()
	client.On("ListSilenceds", mock.Anything, "", "", mock.Anything).Return([]types.Silenced{
		*types.FixtureSilenced("*:my-check"),
		*types.FixtureSilenced("unix:not-my-check"),
	}, nil).Once()

	impl := &checkCfgImpl{}
	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = check

	// return associated silence
	res, err := impl.IsSilenced(params)
	require.NoError(t, err)
	assert.True(t, res)
}

func TestCheckConfigTypeSilencesField(t *testing.T) {
	check := types.FixtureCheckConfig("my-check")
	check.Subscriptions = []string{"unix"}

	client, _ := client.NewClientFactory()
	client.On("ListSilenceds", mock.Anything, "", "", mock.Anything).Return([]types.Silenced{
		*types.FixtureSilenced("*:my-check"),
		*types.FixtureSilenced("unix:*"),
		*types.FixtureSilenced("unix:my-check"),
		*types.FixtureSilenced("unix:different-check"),
		*types.FixtureSilenced("unrelated:my-check"),
		*types.FixtureSilenced("*:another-check"),
	}, nil).Once()

	impl := &checkCfgImpl{}
	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = check

	// return associated silence
	res, err := impl.Silences(params)
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckConfigTypeRuntimeAssetsField(t *testing.T) {
	check := types.FixtureCheckConfig("my-check")
	check.RuntimeAssets = []string{"one", "two"}

	client, _ := client.NewClientFactory()
	client.On("ListAssets", mock.Anything, mock.Anything).Return([]types.Asset{
		*types.FixtureAsset("one"),
		*types.FixtureAsset("two"),
		*types.FixtureAsset("three"),
	}, nil).Once()

	// return associated silence
	impl := &checkCfgImpl{}
	ctx := contextWithLoaders(context.Background(), client)
	res, err := impl.RuntimeAssets(graphql.ResolveParams{Source: check, Context: ctx})
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckConfigTypeHandlersField(t *testing.T) {
	check := types.FixtureCheckConfig("my-check")
	check.Handlers = []string{"one", "two"}

	client, _ := client.NewClientFactory()
	impl := &checkCfgImpl{}

	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = check

	// return associated silence
	client.On("ListHandlers", mock.Anything, mock.Anything).Return([]types.Handler{
		*types.FixtureHandler("one"),
		*types.FixtureHandler("two"),
		*types.FixtureHandler("three"),
	}, nil).Once()

	res, err := impl.Handlers(params)
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckTypeHandlersField(t *testing.T) {
	check := types.FixtureCheck("my-check")
	check.Handlers = []string{"one", "two"}

	client, _ := client.NewClientFactory()
	impl := &checkImpl{}

	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = check

	// return associated silence
	client.On("ListHandlers", mock.Anything, mock.Anything).Return([]types.Handler{
		*types.FixtureHandler("one"),
		*types.FixtureHandler("two"),
		*types.FixtureHandler("three"),
	}, nil).Once()

	res, err := impl.Handlers(params)
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckConfigTypeOutputMetricHandlersField(t *testing.T) {
	check := types.FixtureCheckConfig("my-check")
	check.OutputMetricHandlers = []string{"one", "two"}

	client, _ := client.NewClientFactory()
	impl := &checkCfgImpl{}

	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = check

	// return associated silence
	client.On("ListHandlers", mock.Anything, mock.Anything).Return([]types.Handler{
		*types.FixtureHandler("one"),
		*types.FixtureHandler("two"),
		*types.FixtureHandler("three"),
	}, nil).Once()

	res, err := impl.OutputMetricHandlers(params)
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckTypeOutputMetricHandlersField(t *testing.T) {
	check := types.FixtureCheck("my-check")
	check.OutputMetricHandlers = []string{"one", "two"}

	client, _ := client.NewClientFactory()
	impl := &checkImpl{}

	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = check

	// return associated silence
	client.On("ListHandlers", mock.Anything, mock.Anything).Return([]types.Handler{
		*types.FixtureHandler("one"),
		*types.FixtureHandler("two"),
		*types.FixtureHandler("three"),
	}, nil).Once()

	res, err := impl.OutputMetricHandlers(params)
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckTypeToJSONField(t *testing.T) {
	src := v2.FixtureCheck("name")
	imp := &checkImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestCheckConfigTypeToJSONField(t *testing.T) {
	src := v2.FixtureCheckConfig("name")
	imp := &checkCfgImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}
