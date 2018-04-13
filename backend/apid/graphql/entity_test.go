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
	mock := mockEntityQuerier{els: []*types.Entity{
		types.FixtureEntity("a"),
		types.FixtureEntity("b"),
	}}
	impl := entityImpl{entityCtrl: mock}

	params := schema.EntityRelatedFieldResolverParams{}
	params.Source = types.FixtureEntity("c")
	params.Args.Limit = 10

	res, err := impl.Related(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)

}
