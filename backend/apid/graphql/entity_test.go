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
	source := types.FixtureEntity("c")
	mockCtrl := mockEntityQuerier{els: []*types.Entity{
		source,
		types.FixtureEntity("a"),
		types.FixtureEntity("b"),
	}}

	params := schema.EntityRelatedFieldResolverParams{}
	params.Source = source
	params.Args.Limit = 10

	impl := entityImpl{entityQuerier: mockCtrl}
	res, err := impl.Related(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.Len(t, res, 2)
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
	impl := &entityImpl{eventQuerier: mock}
	st, err := impl.Status(params)
	require.NoError(t, err)
	assert.Equal(t, 0, st)

	// Add failing event
	failingEv := types.FixtureEvent("a", entity.ID)
	failingEv.Check.Status = 2
	mock.els = append(mock.els, failingEv)

	// exit status: 2
	impl = &entityImpl{eventQuerier: mock}
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

func TestEntityTypeEventsField(t *testing.T) {
	entity := types.FixtureEntity("en")
	mock := mockEventQuerier{els: []*types.Event{
		types.FixtureEvent("a", entity.ID),
		types.FixtureEvent("b", entity.ID),
		types.FixtureEvent("c", entity.ID),
	}}

	// params
	params := schema.EntityEventsFieldResolverParams{}
	params.Source = entity

	// return all events
	impl := &entityImpl{eventQuerier: mock}
	evs, err := impl.Events(params)
	require.NoError(t, err)
	assert.Len(t, evs, 3)
}
