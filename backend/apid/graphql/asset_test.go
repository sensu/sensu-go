package graphql

import (
	"context"
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetTypeToJSONField(t *testing.T) {
	src := v2.FixtureAsset("name")
	imp := &assetImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src, Context: context.Background()})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}
