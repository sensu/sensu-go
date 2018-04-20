package graphql

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEntityQuerier struct {
	els []*types.Entity
	err error
}

func (e mockEntityQuerier) Query(ctx context.Context) ([]*types.Entity, error) {
	return e.els, e.err
}

func TestEntityTypeRelatedField(t *testing.T) {
	source := types.FixtureEntity("c")
	mockCtrl := mockEntityQuerier{els: []*types.Entity{
		source,
		types.FixtureEntity("a"),
		types.FixtureEntity("b"),
	}}

	params := schema.EntityRelatedFieldResolverParams{}
	params.Source = source
	params.Args.Limit = 10

	impl := entityImpl{entityCtrl: mockCtrl}
	res, err := impl.Related(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.Len(t, res, 2)
}
