package graphql

import (
	"context"
	"testing"

	v2 "github.com/sensu/sensu-go/api/core/v2"
	client "github.com/sensu/sensu-go/backend/apid/graphql/mockclient"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEventFilterTypeRuntimeAssetsField(t *testing.T) {
	filter := types.FixtureEventFilter("my-filter")
	filter.RuntimeAssets = []string{"one", "two"}

	client, _ := client.NewClientFactory()
	client.On("ListAssets", mock.Anything, mock.Anything).Return([]types.Asset{
		*types.FixtureAsset("one"),
		*types.FixtureAsset("two"),
		*types.FixtureAsset("three"),
	}, nil).Once()

	// return associated silence
	impl := &eventFilterImpl{}
	ctx := contextWithLoaders(context.Background(), client)
	res, err := impl.RuntimeAssets(graphql.ResolveParams{Source: filter, Context: ctx})
	require.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestEventFilterTypeToJSONField(t *testing.T) {
	src := v2.FixtureEventFilter("my-filter")
	imp := &eventFilterImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}
