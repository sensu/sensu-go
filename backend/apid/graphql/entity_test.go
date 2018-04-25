package graphql

import (
	"context"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
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

func TestEntityTypeStatusField(t *testing.T) {
	entity := types.FixtureEntity("en")
	mock := mockEventQuerier{els: []*types.Event{
		types.FixtureEvent("a", entity.ID),
		types.FixtureEvent("b", entity.ID),
		types.FixtureEvent("c", entity.ID),
	}}

	// params
	params := graphql.ResolveParams{}
	params.Source = entity

	// exit status: 0
	impl := &entityImpl{eventCtrl: mock}
	st, err := impl.Status(params)
	require.NoError(t, err)
	assert.Equal(t, 0, st)

	// Add failing event
	failingEv := types.FixtureEvent("a", entity.ID)
	failingEv.Check.Status = 2
	mock.els = append(mock.els, failingEv)

	// exit status: 2
	impl = &entityImpl{eventCtrl: mock}
	st, err = impl.Status(params)
	require.NoError(t, err)
	assert.Equal(t, 2, st)
}

func TestEntityTypeLastSeenField(t *testing.T) {
	now := time.Now()

	entity := types.FixtureEntity("id")
	entity.LastSeen = now.Unix()
	params := graphql.ResolveParams{}
	params.Source = entity

	impl := entityImpl{}
	res, err := impl.LastSeen(params)
	require.NoError(t, err)
	require.NotEmpty(t, res)
	assert.Equal(t, res.Unix(), now.Unix())
}
