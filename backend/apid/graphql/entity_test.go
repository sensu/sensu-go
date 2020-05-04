package graphql

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEntityTypeMetadataField(t *testing.T) {
	src := corev2.FixtureEntity("bug")
	impl := entityImpl{}

	res, err := impl.Metadata(graphql.ResolveParams{Source: src})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.IsType(t, v2.ObjectMeta{}, res)
}

func TestEntityTypeRelatedField(t *testing.T) {
	source := corev2.FixtureEntity("c")

	client := new(MockEntityClient)
	client.On("ListEntities", mock.Anything).Return([]*corev2.Entity{
		source,
		corev2.FixtureEntity("a"),
		corev2.FixtureEntity("b"),
	}, nil).Once()

	cfg := ServiceConfig{EntityClient: client}
	params := schema.EntityRelatedFieldResolverParams{}
	params.Context = contextWithLoaders(context.Background(), cfg)
	params.Source = source
	params.Args.Limit = 10

	impl := entityImpl{}
	res, err := impl.Related(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.Len(t, res, 2)
}

func TestEntityTypeStatusField(t *testing.T) {
	entity := corev2.FixtureEntity("en")
	entity.Namespace = "sensu"

	client := new(MockEventClient)
	client.On("ListEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{
		corev2.FixtureEvent(entity.Name, "a"),
		corev2.FixtureEvent(entity.Name, "b"),
		corev2.FixtureEvent(entity.Name, "c"),
	}, nil).Once()

	// params
	params := graphql.ResolveParams{}
	cfg := ServiceConfig{EventClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = entity

	// exit status: 0
	impl := &entityImpl{}
	st, err := impl.Status(params)
	require.NoError(t, err)
	assert.EqualValues(t, 0, st)

	// Add failing event
	failingEv := corev2.FixtureEvent(entity.Name, "bad")
	failingEv.Check.Status = 2
	client.On("ListEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{
		corev2.FixtureEvent(entity.Name, "a"),
		failingEv,
	}, nil).Once()

	// exit status: 2
	// params.Context = contextWithLoaders(context.Background(), client)
	st, err = impl.Status(params)
	require.NoError(t, err)
	assert.EqualValues(t, 2, st)
}

func TestEntityTypeLastSeenField(t *testing.T) {
	now := time.Now()

	entity := corev2.FixtureEntity("id")
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
	entity := corev2.FixtureEntity("en")

	client := new(MockEventClient)
	client.On("ListEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{
		corev2.FixtureEvent(entity.Name, "a"),
		corev2.FixtureEvent(entity.Name, "b"),
		corev2.FixtureEvent("no-entity", "c"),
	}, nil).Once()

	// params
	params := schema.EntityEventsFieldResolverParams{}
	cfg := ServiceConfig{EventClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Args.Filters = []string{}
	params.Source = entity

	// return all events
	impl := &entityImpl{}
	evs, err := impl.Events(params)
	require.NoError(t, err)
	assert.Len(t, evs, 2)
}

func TestEntityTypeSilencesField(t *testing.T) {
	entity := corev2.FixtureEntity("en")
	entity.Subscriptions = []string{"entity:en", "unix", "www"}

	client := new(MockSilencedClient)
	client.On("ListSilenced", mock.Anything).Return([]*corev2.Silenced{
		corev2.FixtureSilenced("entity:en:*"),
		corev2.FixtureSilenced("www:*"),
		corev2.FixtureSilenced("unix:my-check"),
		corev2.FixtureSilenced("entity:unrelated:*"),
	}, nil).Once()

	impl := &entityImpl{}
	params := graphql.ResolveParams{}
	cfg := ServiceConfig{SilencedClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = entity

	// return associated silence
	evs, err := impl.Silences(params)
	require.NoError(t, err)
	assert.Len(t, evs, 2)
}

func TestEntityTypeIsSilencedField(t *testing.T) {
	entity := corev2.FixtureEntity("en")
	entity.Subscriptions = []string{"entity:en", "ou"}

	client := new(MockSilencedClient)
	client.On("ListSilenced", mock.Anything).Return([]*corev2.Silenced{
		corev2.FixtureSilenced("entity:en:*"),
		corev2.FixtureSilenced("ou:my-check"),
		corev2.FixtureSilenced("entity:unrelated:*"),
	}, nil).Once()

	impl := &entityImpl{}
	params := graphql.ResolveParams{}
	cfg := ServiceConfig{SilencedClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = entity

	// return associated silence
	res, err := impl.IsSilenced(params)
	require.NoError(t, err)
	assert.True(t, res)
}

func TestEntityTypeToJSONField(t *testing.T) {
	src := corev2.FixtureEntity("name")
	imp := &entityImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func Test_processImpl_Created(t *testing.T) {
	src := &corev2.Process{Created: 1588381473555}
	imp := &processImpl{}

	res, err := imp.Created(graphql.ResolveParams{Source: src})
	require.NoError(t, err)
	assert.Equal(t, res, time.Unix(1588381473, 555000000))
}

func Test_processImpl_MemoryPercent(t *testing.T) {
	src := &corev2.Process{MemoryPercent: 1.5}
	imp := &processImpl{}

	res, err := imp.MemoryPercent(graphql.ResolveParams{Source: src})
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprint(res), "1.5")
}
