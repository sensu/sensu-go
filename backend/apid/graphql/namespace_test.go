package graphql

import (
	"context"
	"errors"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNamespaceTypeColourID(t *testing.T) {
	impl := &namespaceImpl{}
	nsp := corev2.Namespace{Name: "sensu"}

	colour, err := impl.ColourID(graphql.ResolveParams{Source: &nsp})
	assert.NoError(t, err)
	assert.Equal(t, string(colour), "ORANGE")
}

func TestNamespaceTypeCheckConfigsField(t *testing.T) {
	checkClient := new(MockCheckClient)
	checkClient.On("ListChecks", mock.Anything).Return([]*corev2.CheckConfig{
		corev2.FixtureCheckConfig("a"),
		corev2.FixtureCheckConfig("b"),
		corev2.FixtureCheckConfig("c"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceChecksFieldResolverParams{}
	cfg := ServiceConfig{CheckClient: checkClient}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev2.FixtureNamespace("default")
	params.Args.Limit = 20

	// Success
	res, err := impl.Checks(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.(offsetContainer).Nodes)

	// Store err
	checkClient.On("ListChecks", mock.Anything).Return([]*corev2.CheckConfig{}, errors.New("abc")).Once()
	res, err = impl.Checks(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeEntitiesField(t *testing.T) {
	client := new(MockEntityClient)
	client.On("ListEntities", mock.Anything).Return([]*corev2.Entity{
		corev2.FixtureEntity("a"),
		corev2.FixtureEntity("b"),
		corev2.FixtureEntity("c"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceEntitiesFieldResolverParams{}
	cfg := ServiceConfig{EntityClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev2.FixtureNamespace("default")
	params.Args.Limit = 20

	// Success
	res, err := impl.Entities(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.(offsetContainer).Nodes)

	// Store err
	client.On("ListEntities", mock.Anything, mock.Anything).Return([]*corev2.Entity{}, errors.New("abc")).Once()
	res, err = impl.Entities(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeEventsField(t *testing.T) {
	client := new(MockEventClient)
	client.On("ListEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{
		corev2.FixtureEvent("a", "b"),
		corev2.FixtureEvent("b", "c"),
		corev2.FixtureEvent("c", "d"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceEventsFieldResolverParams{}
	cfg := ServiceConfig{EventClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev2.FixtureNamespace("default")
	params.Args.Limit = 20

	// Success
	res, err := impl.Events(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.(offsetContainer).Nodes)

	// Store err
	client.On("ListEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{}, errors.New("abc")).Once()
	res, err = impl.Events(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeEventFiltersField(t *testing.T) {
	client := new(MockEventFilterClient)
	client.On("ListEventFilters", mock.Anything).Return([]*corev2.EventFilter{
		corev2.FixtureEventFilter("a"),
		corev2.FixtureEventFilter("b"),
		corev2.FixtureEventFilter("c"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceEventFiltersFieldResolverParams{}
	cfg := ServiceConfig{EventFilterClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev2.FixtureNamespace("default")
	params.Args.Limit = 20

	// Success
	res, err := impl.EventFilters(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.(offsetContainer).Nodes)

	// Store err
	client.On("ListEventFilters", mock.Anything, mock.Anything).Return([]*corev2.EventFilter{}, errors.New("abc")).Once()
	res, err = impl.EventFilters(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeHandlersField(t *testing.T) {
	client := new(MockHandlerClient)
	client.On("ListHandlers", mock.Anything).Return([]*corev2.Handler{
		corev2.FixtureHandler("Abe"),
		corev2.FixtureHandler("Bernie"),
		corev2.FixtureHandler("Clem"),
		corev2.FixtureHandler("Dolly"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceHandlersFieldResolverParams{}
	cfg := ServiceConfig{HandlerClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev2.FixtureNamespace("default")
	params.Args.Limit = 10

	// Success
	res, err := impl.Handlers(params)
	require.NoError(t, err)
	assert.Len(t, res.(offsetContainer).Nodes, 4)

	// Store err
	client.On("ListHandlers", mock.Anything, mock.Anything).Return(
		[]*corev2.Handler{},
		errors.New("error"),
	)

	res, err = impl.Handlers(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeMutatorsField(t *testing.T) {
	client := new(MockMutatorClient)
	client.On("ListMutators", mock.Anything).Return([]*corev2.Mutator{
		corev2.FixtureMutator("Abe"),
		corev2.FixtureMutator("Bernie"),
		corev2.FixtureMutator("Clem"),
		corev2.FixtureMutator("Dolly"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceMutatorsFieldResolverParams{}
	cfg := ServiceConfig{MutatorClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev2.FixtureNamespace("default")
	params.Args.Limit = 10

	// Success
	res, err := impl.Mutators(params)
	require.NoError(t, err)
	assert.Len(t, res.(offsetContainer).Nodes, 4)

	// Store err
	client.On("ListMutators", mock.Anything, mock.Anything).Return(
		[]*corev2.Mutator{},
		errors.New("error"),
	)

	res, err = impl.Mutators(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeSilencesField(t *testing.T) {
	client := new(MockSilencedClient)
	client.On("ListSilenced", mock.Anything).Return([]*corev2.Silenced{
		corev2.FixtureSilenced("a:b"),
		corev2.FixtureSilenced("b:c"),
		corev2.FixtureSilenced("c:d"),
	}, nil).Once()

	impl := &namespaceImpl{}
	cfg := ServiceConfig{SilencedClient: client}
	params := schema.NamespaceSilencesFieldResolverParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev2.FixtureNamespace("xxx")

	// Success
	res, err := impl.Silences(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)

	// Store err
	client.On("ListSilenced", mock.Anything).Return([]*corev2.Silenced{}, errors.New("abc"))
	res, err = impl.Silences(params)
	assert.Empty(t, res.(offsetContainer).Nodes)
	assert.Error(t, err)
}
