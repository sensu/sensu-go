package graphql

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	client "github.com/sensu/sensu-go/backend/apid/graphql/mockclient"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEventFilterTypeRuntimeAssetsField(t *testing.T) {
	filter := corev2.FixtureEventFilter("my-filter")
	filter.RuntimeAssets = []string{"one", "two"}

	_, factory := client.NewClientFactory()
	assetClient := new(MockAssetClient)
	assetClient.On("ListAssets", mock.Anything, mock.Anything).Return([]*corev2.Asset{
		corev2.FixtureAsset("one"),
		corev2.FixtureAsset("two"),
		corev2.FixtureAsset("three"),
	}, nil).Once()

	// return associated silence
	impl := &eventFilterImpl{}
	cfg := ServiceConfig{ClientFactory: factory, AssetClient: assetClient}
	ctx := contextWithLoaders(context.Background(), cfg)
	res, err := impl.RuntimeAssets(graphql.ResolveParams{Source: filter, Context: ctx})
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestEventFilterTypeToJSONField(t *testing.T) {
	src := corev2.FixtureEventFilter("my-filter")
	imp := &eventFilterImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}
