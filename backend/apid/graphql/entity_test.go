package graphql

import (
	"context"
	"testing"
	"time"

	client "github.com/sensu/sensu-go/backend/apid/graphql/mockclient"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEntityTypeRelatedField(t *testing.T) {
	source := types.FixtureEntity("c")

	client, _ := client.NewClientFactory()
	client.On("ListEntities", mock.Anything, mock.Anything).Return([]types.Entity{
		*source,
		*types.FixtureEntity("a"),
		*types.FixtureEntity("b"),
	}, nil).Once()

	params := schema.EntityRelatedFieldResolverParams{}
	params.Context = contextWithLoaders(context.Background(), client)
	params.Source = source
	params.Args.Limit = 10

	impl := entityImpl{}
	res, err := impl.Related(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.Len(t, res, 2)
}

func TestEntityTypeStatusField(t *testing.T) {
	entity := types.FixtureEntity("en")
	entity.Namespace = "sensu"

	client, _ := client.NewClientFactory()
	client.On("ListEvents", "sensu", mock.Anything).Return([]types.Event{
		*types.FixtureEvent(entity.Name, "a"),
		*types.FixtureEvent(entity.Name, "b"),
		*types.FixtureEvent(entity.Name, "c"),
	}, nil).Once()

	// params
	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = entity

	// exit status: 0
	impl := &entityImpl{}
	st, err := impl.Status(params)
	require.NoError(t, err)
	assert.EqualValues(t, 0, st)

	// Add failing event
	failingEv := types.FixtureEvent(entity.Name, "bad")
	failingEv.Check.Status = 2
	client.On("ListEvents", "sensu", mock.Anything).Return([]types.Event{
		*types.FixtureEvent(entity.Name, "a"),
		*failingEv,
	}, nil).Once()

	// exit status: 2
	// params.Context = contextWithLoaders(context.Background(), client)
	st, err = impl.Status(params)
	require.NoError(t, err)
	assert.EqualValues(t, 2, st)
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

	client, _ := client.NewClientFactory()
	client.On("ListEvents", mock.Anything, mock.Anything).Return([]types.Event{
		*types.FixtureEvent(entity.Name, "a"),
		*types.FixtureEvent(entity.Name, "b"),
		*types.FixtureEvent(entity.Name, "c"),
	}, nil).Once()

	// params
	params := schema.EntityEventsFieldResolverParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = entity

	// return all events
	impl := &entityImpl{}
	evs, err := impl.Events(params)
	require.NoError(t, err)
	assert.Len(t, evs, 3)
}

func TestEntityTypeSilencesField(t *testing.T) {
	entity := types.FixtureEntity("en")
	entity.Subscriptions = []string{"entity:en", "unix", "www"}

	client, _ := client.NewClientFactory()
	client.On("ListSilenceds", mock.Anything, "", "").Return([]types.Silenced{
		*types.FixtureSilenced("entity:en:*"),
		*types.FixtureSilenced("www:*"),
		*types.FixtureSilenced("unix:my-check"),
		*types.FixtureSilenced("entity:unrelated:*"),
	}, nil).Once()

	impl := &entityImpl{}
	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = entity

	// return associated silence
	evs, err := impl.Silences(params)
	require.NoError(t, err)
	assert.Len(t, evs, 2)
}

func TestEntityTypeIsSilencedField(t *testing.T) {
	entity := types.FixtureEntity("en")
	entity.Subscriptions = []string{"entity:en", "ou"}

	client, _ := client.NewClientFactory()
	client.On("ListSilenceds", mock.Anything, "", "").Return([]types.Silenced{
		*types.FixtureSilenced("entity:en:*"),
		*types.FixtureSilenced("ou:my-check"),
		*types.FixtureSilenced("entity:unrelated:*"),
	}, nil).Once()

	impl := &entityImpl{}
	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = entity

	// return associated silence
	res, err := impl.IsSilenced(params)
	require.NoError(t, err)
	assert.True(t, res)
}
