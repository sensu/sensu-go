package graphql

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEventFilterTypeRuntimeAssetsField(t *testing.T) {
	filter := corev2.FixtureEventFilter("my-filter")
	filter.RuntimeAssets = []string{"one", "two", "four", "six:seven"}

	assetClient := new(MockAssetClient)
	assetClient.On("ListAssets", mock.Anything, mock.Anything).Return([]*corev2.Asset{
		corev2.FixtureAsset("one"),
		corev2.FixtureAsset("two"),
		corev2.FixtureAsset("three"),
		corev2.FixtureAsset("four:five"),
		corev2.FixtureAsset("six:seven"),
	}, nil).Once()

	// return associated silence
	impl := &eventFilterImpl{}
	cfg := ServiceConfig{AssetClient: assetClient}
	ctx := contextWithLoaders(context.Background(), cfg)
	res, err := impl.RuntimeAssets(graphql.ResolveParams{Source: filter, Context: ctx})
	require.NoError(t, err)
	assert.Len(t, res, 3)
}

func TestEventFilterTypeToJSONField(t *testing.T) {
	src := corev2.FixtureEventFilter("my-filter")
	imp := &eventFilterImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src, Context: context.Background()})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}
