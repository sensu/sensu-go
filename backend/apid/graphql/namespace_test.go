package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNamespaceTypeColourID(t *testing.T) {
	impl := &namespaceImpl{}
	nsp := types.Namespace{Name: "sensu"}

	colour, err := impl.ColourID(graphql.ResolveParams{Source: &nsp})
	assert.NoError(t, err)
	assert.Equal(t, string(colour), "ORANGE")
}

func TestNamespaceTypeCheckConfigsField(t *testing.T) {
	checkClient := new(MockCheckClient)
	checkClient.On("ListChecks", mock.Anything).Return([]*types.CheckConfig{
		types.FixtureCheckConfig("a"),
		types.FixtureCheckConfig("b"),
		types.FixtureCheckConfig("c"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceChecksFieldResolverParams{}
	cfg := ServiceConfig{CheckClient: checkClient}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = types.FixtureNamespace("default")
	params.Args.Limit = 20

	// Success
	res, err := impl.Checks(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.(offsetContainer).Nodes)

	// Store err
	checkClient.On("ListChecks", mock.Anything).Return([]*types.CheckConfig{}, errors.New("abc")).Once()
	res, err = impl.Checks(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeEntitiesField(t *testing.T) {
	client := new(MockEntityClient)
	client.On("ListEntities", mock.Anything).Return([]*types.Entity{
		types.FixtureEntity("a"),
		types.FixtureEntity("b"),
		types.FixtureEntity("c"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceEntitiesFieldResolverParams{}
	cfg := ServiceConfig{EntityClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = types.FixtureNamespace("default")
	params.Args.Limit = 20

	// Success
	res, err := impl.Entities(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.(offsetContainer).Nodes)

	// Store err
	client.On("ListEntities", mock.Anything, mock.Anything).Return([]*types.Entity{}, errors.New("abc")).Once()
	res, err = impl.Entities(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeEventsField(t *testing.T) {
	client := new(MockEventClient)
	client.On("ListEvents", mock.Anything, mock.Anything).Return([]*types.Event{
		types.FixtureEvent("a", "b"),
		types.FixtureEvent("b", "c"),
		types.FixtureEvent("c", "d"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceEventsFieldResolverParams{}
	cfg := ServiceConfig{EventClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = types.FixtureNamespace("default")
	params.Args.Limit = 20

	// Success
	res, err := impl.Events(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.(offsetContainer).Nodes)

	// Store err
	client.On("ListEvents", mock.Anything, mock.Anything).Return([]*types.Event{}, errors.New("abc")).Once()
	res, err = impl.Events(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeEventFiltersField(t *testing.T) {
	client := new(MockEventFilterClient)
	client.On("ListEventFilters", mock.Anything).Return([]*types.EventFilter{
		types.FixtureEventFilter("a"),
		types.FixtureEventFilter("b"),
		types.FixtureEventFilter("c"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceEventFiltersFieldResolverParams{}
	cfg := ServiceConfig{EventFilterClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = types.FixtureNamespace("default")
	params.Args.Limit = 20

	// Success
	res, err := impl.EventFilters(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.(offsetContainer).Nodes)

	// Store err
	client.On("ListEventFilters", mock.Anything, mock.Anything).Return([]*types.EventFilter{}, errors.New("abc")).Once()
	res, err = impl.EventFilters(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeHandlersField(t *testing.T) {
	client := new(MockHandlerClient)
	client.On("ListHandlers", mock.Anything).Return([]*types.Handler{
		types.FixtureHandler("Abe"),
		types.FixtureHandler("Bernie"),
		types.FixtureHandler("Clem"),
		types.FixtureHandler("Dolly"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceHandlersFieldResolverParams{}
	cfg := ServiceConfig{HandlerClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = types.FixtureNamespace("default")
	params.Args.Limit = 10

	// Success
	res, err := impl.Handlers(params)
	require.NoError(t, err)
	assert.Len(t, res.(offsetContainer).Nodes, 4)

	// Store err
	client.On("ListHandlers", mock.Anything, mock.Anything).Return(
		[]*types.Handler{},
		errors.New("error"),
	)

	res, err = impl.Handlers(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeMutatorsField(t *testing.T) {
	client := new(MockMutatorClient)
	client.On("ListMutators", mock.Anything).Return([]*types.Mutator{
		types.FixtureMutator("Abe"),
		types.FixtureMutator("Bernie"),
		types.FixtureMutator("Clem"),
		types.FixtureMutator("Dolly"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceMutatorsFieldResolverParams{}
	cfg := ServiceConfig{MutatorClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = types.FixtureNamespace("default")
	params.Args.Limit = 10

	// Success
	res, err := impl.Mutators(params)
	require.NoError(t, err)
	assert.Len(t, res.(offsetContainer).Nodes, 4)

	// Store err
	client.On("ListMutators", mock.Anything, mock.Anything).Return(
		[]*types.Mutator{},
		errors.New("error"),
	)

	res, err = impl.Mutators(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeSilencesField(t *testing.T) {
	client := new(MockSilencedClient)
	client.On("ListSilenced", mock.Anything).Return([]*types.Silenced{
		types.FixtureSilenced("a:b"),
		types.FixtureSilenced("b:c"),
		types.FixtureSilenced("c:d"),
	}, nil).Once()

	impl := &namespaceImpl{}
	cfg := ServiceConfig{SilencedClient: client}
	params := schema.NamespaceSilencesFieldResolverParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = types.FixtureNamespace("xxx")

	// Success
	res, err := impl.Silences(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)

	// Store err
	client.On("ListSilenced", mock.Anything).Return([]*types.Silenced{}, errors.New("abc"))
	res, err = impl.Silences(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}
