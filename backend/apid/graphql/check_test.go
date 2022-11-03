package graphql

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
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

	check := corev2.FixtureCheck("test")
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("w/ argument of %d", tc.expectedLen), func(t *testing.T) {
			params := schema.CheckHistoryFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
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
	check := corev2.FixtureCheck("test")
	check.LastOK = now.Unix()

	impl := checkImpl{}
	params := graphql.ResolveParams{Source: check}

	res, err := impl.LastOK(params)
	require.NoError(t, err)
	assert.Equal(t, now.Unix(), res.Unix())
}

func TestCheckTypeIssuedFieldImpl(t *testing.T) {
	now := time.Now()
	check := corev2.FixtureCheck("test")
	check.Issued = now.Unix()

	impl := checkImpl{}
	params := graphql.ResolveParams{Source: check}

	res, err := impl.Issued(params)
	require.NoError(t, err)
	assert.Equal(t, now.Unix(), res.Unix())
}

func TestCheckTypeNodeIDFieldImpl(t *testing.T) {
	check := corev2.FixtureCheck("test")
	params := graphql.ResolveParams{Source: check}

	impl := checkImpl{}
	res, err := impl.NodeID(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestCheckTypeIsSilencedField(t *testing.T) {
	check := corev2.FixtureCheck("my-check")
	check.Silenced = []string{"unix:my-check"}

	// return associated silence
	impl := &checkImpl{}
	res, err := impl.IsSilenced(graphql.ResolveParams{Source: check, Context: context.Background()})
	require.NoError(t, err)
	assert.True(t, res)
}

func TestCheckTypeSilencesField(t *testing.T) {
	check := corev2.FixtureCheck("my-check")
	check.Subscriptions = []string{"unix"}
	check.Silenced = []string{"unix:my-check"}

	client := new(MockSilencedClient)
	client.On("ListSilenced", mock.Anything).Return([]*corev2.Silenced{
		corev2.FixtureSilenced("unix:my-check"),
		corev2.FixtureSilenced("fred:my-check"),
		corev2.FixtureSilenced("unix:not-my-check"),
	}, nil).Once()

	impl := &checkImpl{}
	cfg := ServiceConfig{SilencedClient: client}
	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = check

	// return associated silence
	res, err := impl.Silences(params)
	require.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestCheckTypeRuntimeAssetsField(t *testing.T) {
	check := corev2.FixtureCheck("my-check")
	check.RuntimeAssets = []string{"one", "two"}

	assetClient := new(MockAssetClient)
	assetClient.On("ListAssets", mock.Anything).Return([]*corev2.Asset{
		corev2.FixtureAsset("one"),
		corev2.FixtureAsset("two"),
		corev2.FixtureAsset("three"),
	}, nil).Once()

	// return associated silence
	impl := &checkImpl{}
	cfg := ServiceConfig{AssetClient: assetClient}
	ctx := contextWithLoaders(context.Background(), cfg)
	res, err := impl.RuntimeAssets(graphql.ResolveParams{Source: check, Context: ctx})
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckConfigTypeIsSilencedField(t *testing.T) {
	check := corev2.FixtureCheckConfig("my-check")
	check.Subscriptions = []string{"unix"}

	client := new(MockSilencedClient)
	client.On("ListSilenced", mock.Anything).Return([]*corev2.Silenced{
		corev2.FixtureSilenced("*:my-check"),
		corev2.FixtureSilenced("unix:not-my-check"),
	}, nil).Once()

	impl := &checkCfgImpl{}
	params := graphql.ResolveParams{}
	cfg := ServiceConfig{SilencedClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = check

	// return associated silence
	res, err := impl.IsSilenced(params)
	require.NoError(t, err)
	assert.True(t, res)
}

func TestCheckConfigTypeSilencesField(t *testing.T) {
	check := corev2.FixtureCheckConfig("my-check")
	check.Subscriptions = []string{"unix"}

	client := new(MockSilencedClient)
	client.On("ListSilenced", mock.Anything).Return([]*corev2.Silenced{
		corev2.FixtureSilenced("*:my-check"),
		corev2.FixtureSilenced("unix:*"),
		corev2.FixtureSilenced("unix:my-check"),
		corev2.FixtureSilenced("unix:different-check"),
		corev2.FixtureSilenced("unrelated:my-check"),
		corev2.FixtureSilenced("*:another-check"),
	}, nil).Once()

	impl := &checkCfgImpl{}
	params := graphql.ResolveParams{}
	cfg := ServiceConfig{SilencedClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = check

	// return associated silence
	res, err := impl.Silences(params)
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckConfigTypeRuntimeAssetsField(t *testing.T) {
	check := corev2.FixtureCheckConfig("my-check")
	check.RuntimeAssets = []string{"one", "two"}

	assetClient := new(MockAssetClient)
	assetClient.On("ListAssets", mock.Anything).Return([]*corev2.Asset{
		corev2.FixtureAsset("one"),
		corev2.FixtureAsset("two"),
		corev2.FixtureAsset("three"),
	}, nil).Once()

	// return associated silence
	impl := &checkCfgImpl{}
	cfg := ServiceConfig{AssetClient: assetClient}
	ctx := contextWithLoaders(context.Background(), cfg)
	res, err := impl.RuntimeAssets(graphql.ResolveParams{Source: check, Context: ctx})
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestCheckConfigTypeHandlersField(t *testing.T) {
	check := corev2.FixtureCheckConfig("my-check")
	check.Handlers = []string{"one", "two", "four", "six:seven"}

	impl := &checkCfgImpl{}

	params := graphql.ResolveParams{}
	client := new(MockHandlerClient)

	// return associated silence
	client.On("ListHandlers", mock.Anything).Return([]*corev2.Handler{
		corev2.FixtureHandler("one"),
		corev2.FixtureHandler("two"),
		corev2.FixtureHandler("three"),
		corev2.FixtureHandler("four:five"),
		corev2.FixtureHandler("six:seven"),
	}, nil).Once()

	cfg := ServiceConfig{HandlerClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = check

	res, err := impl.Handlers(params)
	require.NoError(t, err)
	assert.Len(t, res, 3)
}

func TestCheckTypeHandlersField(t *testing.T) {
	check := corev2.FixtureCheck("my-check")
	check.Handlers = []string{"one", "two", "four", "six:seven"}

	client := new(MockHandlerClient)
	impl := &checkImpl{}

	params := graphql.ResolveParams{}
	cfg := ServiceConfig{HandlerClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = check

	// return associated silence
	client.On("ListHandlers", mock.Anything).Return([]*corev2.Handler{
		corev2.FixtureHandler("one"),
		corev2.FixtureHandler("two"),
		corev2.FixtureHandler("three"),
		corev2.FixtureHandler("four:five"),
		corev2.FixtureHandler("six:seven"),
	}, nil).Once()

	res, err := impl.Handlers(params)
	require.NoError(t, err)
	assert.Len(t, res, 3)
}

func TestCheckConfigTypeOutputMetricHandlersField(t *testing.T) {
	check := corev2.FixtureCheckConfig("my-check")
	check.OutputMetricHandlers = []string{"one", "two", "four", "six:seven"}

	client := new(MockHandlerClient)
	impl := &checkCfgImpl{}

	params := graphql.ResolveParams{}
	cfg := ServiceConfig{HandlerClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = check

	// return associated silence
	client.On("ListHandlers", mock.Anything).Return([]*corev2.Handler{
		corev2.FixtureHandler("one"),
		corev2.FixtureHandler("two"),
		corev2.FixtureHandler("three"),
		corev2.FixtureHandler("four:five"),
		corev2.FixtureHandler("six:seven"),
	}, nil).Once()

	res, err := impl.OutputMetricHandlers(params)
	require.NoError(t, err)
	assert.Len(t, res, 3)
}

func TestCheckTypeOutputMetricHandlersField(t *testing.T) {
	check := corev2.FixtureCheck("my-check")
	check.OutputMetricHandlers = []string{"one", "two", "four", "six:seven"}

	client := new(MockHandlerClient)
	impl := &checkImpl{}

	params := graphql.ResolveParams{}
	cfg := ServiceConfig{HandlerClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = check

	// return associated silence
	client.On("ListHandlers", mock.Anything).Return([]*corev2.Handler{
		corev2.FixtureHandler("one"),
		corev2.FixtureHandler("two"),
		corev2.FixtureHandler("three"),
		corev2.FixtureHandler("four:five"),
		corev2.FixtureHandler("six:seven"),
	}, nil).Once()

	res, err := impl.OutputMetricHandlers(params)
	require.NoError(t, err)
	assert.Len(t, res, 3)
}

func TestCheckTypeToJSONField(t *testing.T) {
	src := corev2.FixtureCheck("name")
	imp := &checkImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src, Context: context.Background()})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestCheckConfigTypeToJSONField(t *testing.T) {
	src := corev2.FixtureCheckConfig("name")
	imp := &checkCfgImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src, Context: context.Background()})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestCheckTypeOutputFieldImpl(t *testing.T) {
	testCases := []struct {
		name     string
		result   string
		firstArg int
		lastArg  int
	}{
		{
			name:   "no args",
			result: "123456789012345678901234567890",
		},
		{
			name:     "first 10",
			result:   "1234567890",
			firstArg: 10,
		},
		{
			name:    "last 5",
			result:  "67890",
			lastArg: 5,
		},
		{
			name:     "first 25 & last 10",
			result:   "6789012345",
			firstArg: 25,
			lastArg:  10,
		},
		{
			name:     "first out of bounds",
			result:   "123456789012345678901234567890",
			firstArg: 55,
		},
		{
			name:     "last out of bounds",
			result:   "12345",
			firstArg: 5,
			lastArg:  55,
		},
	}

	check := corev2.FixtureCheck("test")
	check.Output = "123456789012345678901234567890"
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := schema.CheckOutputFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
			params.Context = context.Background()
			params.Source = check
			params.Args.First = tc.firstArg
			params.Args.Last = tc.lastArg

			impl := checkImpl{}
			res, err := impl.Output(params)
			require.NoError(t, err)
			assert.Equal(t, tc.result, res)
		})
	}
}

func Test_checkCfgImpl_IsSilenced(t *testing.T) {
	testCases := []struct {
		name     string
		check    *corev2.CheckConfig
		silence  *corev2.Silenced
		expected bool
	}{
		{
			name:     "matches check",
			check:    corev2.FixtureCheckConfig("check"),
			silence:  corev2.FixtureSilenced("*:check"),
			expected: true,
		},
		{
			name:     "matches subscription",
			check:    &corev2.CheckConfig{Subscriptions: []string{"unix"}},
			silence:  corev2.FixtureSilenced("unix:*"),
			expected: true,
		},
		{
			name:     "no match",
			check:    &corev2.CheckConfig{Subscriptions: []string{"unix"}},
			silence:  corev2.FixtureSilenced("cats:dogs"),
			expected: false,
		},
		{
			name:     "starts in far future",
			check:    corev2.FixtureCheckConfig("check"),
			silence:  &corev2.Silenced{Check: "check", Begin: 999_999_999_999},
			expected: false,
		},
		{
			name:     "starts in distant past",
			check:    corev2.FixtureCheckConfig("check"),
			silence:  &corev2.Silenced{Check: "check", Begin: 0},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := new(MockSilencedClient)
			client.On("ListSilenced", mock.Anything).Return([]*corev2.Silenced{
				tc.silence,
			}, nil).Once()

			impl := &checkCfgImpl{}
			params := graphql.ResolveParams{}
			cfg := ServiceConfig{SilencedClient: client}
			params.Context = contextWithLoadersNoCache(context.Background(), cfg)
			params.Source = tc.check

			// return associated silence
			res, err := impl.IsSilenced(params)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, res)
		})
	}
}
