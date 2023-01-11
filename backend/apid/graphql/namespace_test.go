package graphql

import (
	"context"
	"errors"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNamespaceTypeCheckConfigsField(t *testing.T) {
	checkClient := new(MockCheckClient)
	checkClient.On("ListChecks", mock.Anything).Return([]*corev2.CheckConfig{
		corev2.FixtureCheckConfig("a"),
		corev2.FixtureCheckConfig("b"),
		corev2.FixtureCheckConfig("c"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceChecksFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	cfg := ServiceConfig{CheckClient: checkClient}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev3.FixtureNamespace("default")
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
	client.On("ListEntities", mock.Anything, mock.Anything).Return([]*corev2.Entity{
		corev2.FixtureEntity("a"),
		corev2.FixtureEntity("b"),
		corev2.FixtureEntity("c"),
	}, nil).Times(2)

	params := schema.NamespaceEntitiesFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	params.Context = context.Background()
	params.Source = corev3.FixtureNamespace("default")
	params.Args.Limit = 2

	// Success
	resolver := &namespaceImpl{entityClient: client}
	got, err := resolver.Entities(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, got.(offsetContainer).Nodes)

	// Metrics
	metricsStore := new(MockClusterMetricStore)
	metricsStore.On("EntityCount", mock.Anything, "total").Return(10, nil)
	resolver.serviceConfig = &ServiceConfig{ClusterMetricStore: metricsStore}
	got, err = resolver.Entities(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, got.(offsetContainer).Nodes)
	assert.Equal(t, 10, got.(offsetContainer).PageInfo.totalCount)
	assert.False(t, got.(offsetContainer).PageInfo.partialCount)

	// Multiple chunks
	client.On("ListEntities", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		pred := args[1].(*store.SelectionPredicate)
		pred.Continue = "next"
	}).Return([]*corev2.Entity{
		corev2.FixtureEntity("a"),
		corev2.FixtureEntity("b"),
		corev2.FixtureEntity("c"),
		corev2.FixtureEntity("d"),
		corev2.FixtureEntity("e"),
	}, nil).Times(100)
	params.Args.Limit = 20
	resolver = &namespaceImpl{entityClient: client}
	got, err = resolver.Entities(params)
	assert.NoError(t, err)
	assert.Len(t, got.(offsetContainer).Nodes, 20)
	assert.Equal(t, 500, got.(offsetContainer).PageInfo.totalCount)
	assert.Equal(t, true, got.(offsetContainer).PageInfo.partialCount)

	// Finite chunks
	client.On("ListEntities", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		pred := args[1].(*store.SelectionPredicate)
		pred.Continue = "next"
	}).Return([]*corev2.Entity{
		corev2.FixtureEntity("a"),
		corev2.FixtureEntity("b"),
		corev2.FixtureEntity("c"),
		corev2.FixtureEntity("d"),
		corev2.FixtureEntity("e"),
	}, nil).Times(50)
	client.On("ListEntities", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		pred := args[1].(*store.SelectionPredicate)
		pred.Continue = ""
	}).Return([]*corev2.Entity{
		corev2.FixtureEntity("a"),
		corev2.FixtureEntity("b"),
		corev2.FixtureEntity("c"),
		corev2.FixtureEntity("d"),
		corev2.FixtureEntity("e"),
	}, nil).Once()
	params.Args.Limit = 20
	resolver = &namespaceImpl{entityClient: client}
	got, err = resolver.Entities(params)
	assert.NoError(t, err)
	assert.Len(t, got.(offsetContainer).Nodes, 20)
	assert.Equal(t, 255, got.(offsetContainer).PageInfo.totalCount)
	assert.Equal(t, false, got.(offsetContainer).PageInfo.partialCount)

	// w/ offset
	client.On("ListEntities", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		pred := args[1].(*store.SelectionPredicate)
		pred.Continue = "next"
	}).Return([]*corev2.Entity{
		corev2.FixtureEntity("a"),
		corev2.FixtureEntity("b"),
		corev2.FixtureEntity("c"),
		corev2.FixtureEntity("d"),
		corev2.FixtureEntity("e"),
	}, nil).Times(50)
	client.On("ListEntities", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		pred := args[1].(*store.SelectionPredicate)
		pred.Continue = ""
	}).Return([]*corev2.Entity{
		corev2.FixtureEntity("a"),
		corev2.FixtureEntity("b"),
		corev2.FixtureEntity("c"),
		corev2.FixtureEntity("d"),
		corev2.FixtureEntity("e"),
	}, nil).Once()
	params.Args.Limit = 20
	params.Args.Offset = 250
	resolver = &namespaceImpl{entityClient: client}
	got, err = resolver.Entities(params)
	assert.NoError(t, err)
	assert.Len(t, got.(offsetContainer).Nodes, 5)
	assert.Equal(t, 255, got.(offsetContainer).PageInfo.totalCount)
	assert.Equal(t, false, got.(offsetContainer).PageInfo.partialCount)

	// Store err
	client.On("ListEntities", mock.Anything, mock.Anything).Return([]*corev2.Entity{}, errors.New("abc")).Once()
	got, err = resolver.Entities(params)
	assert.Empty(t, got.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeEventsField(t *testing.T) {
	client := new(MockEventClient)
	client.On("EventStoreSupportsFiltering", mock.Anything).Return(false)
	client.On("ListEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{
		corev2.FixtureEvent("a", "b"),
		corev2.FixtureEvent("b", "c"),
		corev2.FixtureEvent("c", "d"),
	}, nil).Once()

	params := schema.NamespaceEventsFieldResolverParams{}
	params.Context = context.Background()
	params.Source = corev3.FixtureNamespace("default")
	params.Args.Limit = 20

	// Success
	resolver := &namespaceImpl{eventClient: client}
	got, err := resolver.Events(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, got.(offsetContainer).Nodes)

	// Store err
	client.On("ListEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{}, errors.New("abc")).Once()
	got, err = resolver.Events(params)
	assert.Empty(t, got.(offsetContainer).Nodes)
	assert.Error(t, err)
}

func TestNamespaceTypeEventsFieldWithStoreFiltering(t *testing.T) {

	newClient := func(countTotal int64, listEventsErr bool) *MockEventClient {
		client := new(MockEventClient)
		// event client with filtering enabled
		client.On("EventStoreSupportsFiltering", mock.Anything).Return(true)
		client.On("CountEvents", mock.Anything, mock.Anything).Return(countTotal, nil)

		if listEventsErr {
			client.On("ListEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{}, errors.New("abc")).Once()
			return client
		}
		client.On("ListEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{
			corev2.FixtureEvent("a", "b"),
			corev2.FixtureEvent("b", "c"),
			corev2.FixtureEvent("c", "d"),
		}, nil).Once()
		return client
	}

	testCases := []struct {
		name               string
		client             *MockEventClient
		args               schema.NamespaceEventsFieldResolverArgs
		expectErr          bool
		expectedOrdering   string
		expectedDescending bool
		expectedLimit      int64
		expectedOffset     int64
		expectedTotal      int
	}{
		{
			name:   "New query by Newest",
			client: newClient(128, false),
			args: schema.NamespaceEventsFieldResolverArgs{
				Limit:   100,
				OrderBy: schema.EventsListOrders.NEWEST,
			},
			expectedOrdering:   corev2.EventSortTimestamp,
			expectedDescending: true,
			expectedLimit:      100,
			expectedTotal:      128,
		}, {
			name:   "Offset query for Oldest",
			client: newClient(9999, false),
			args: schema.NamespaceEventsFieldResolverArgs{
				Limit:   5,
				Offset:  900,
				OrderBy: schema.EventsListOrders.OLDEST,
			},
			expectedOrdering:   corev2.EventSortTimestamp,
			expectedDescending: false,
			expectedLimit:      5,
			expectedOffset:     900,
			expectedTotal:      9999,
		}, {
			name:   "New query by entity",
			client: newClient(128, false),
			args: schema.NamespaceEventsFieldResolverArgs{
				Limit:   100,
				OrderBy: schema.EventsListOrders.ENTITY,
			},
			expectedOrdering:   corev2.EventSortEntity,
			expectedDescending: false,
			expectedLimit:      100,
			expectedTotal:      128,
		}, {
			name:   "New query by entity descending",
			client: newClient(128, false),
			args: schema.NamespaceEventsFieldResolverArgs{
				Limit:   100,
				OrderBy: schema.EventsListOrders.ENTITY_DESC,
			},
			expectedOrdering:   corev2.EventSortEntity,
			expectedDescending: true,
			expectedLimit:      100,
			expectedTotal:      128,
		}, {
			name:   "New query by last ok",
			client: newClient(128, false),
			args: schema.NamespaceEventsFieldResolverArgs{
				Limit:   100,
				OrderBy: schema.EventsListOrders.LASTOK,
			},
			expectedOrdering:   corev2.EventSortLastOk,
			expectedDescending: true,
			expectedLimit:      100,
			expectedTotal:      128,
		}, {
			name:   "New query ordering not specified",
			client: newClient(128, false),
			args: schema.NamespaceEventsFieldResolverArgs{
				Limit: 100,
			},
			expectedOrdering:   corev2.EventSortLastOk,
			expectedDescending: true,
			expectedLimit:      100,
			expectedTotal:      128,
		}, {
			name:   "Store Error",
			client: newClient(0, true),
			args: schema.NamespaceEventsFieldResolverArgs{
				Limit:   100,
				OrderBy: schema.EventsListOrders.ENTITY_DESC,
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			impl := &namespaceImpl{eventClient: tc.client}
			params := schema.NamespaceEventsFieldResolverParams{
				ResolveParams: graphql.ResolveParams{Context: context.Background()},
				Args:          tc.args,
			}
			params.Context = context.Background()
			params.Source = corev3.FixtureNamespace("default")

			res, err := impl.Events(params)

			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			actual := res.(offsetContainer)
			assert.NotEmpty(t, actual.Nodes)
			assert.Equal(t, tc.expectedTotal, actual.PageInfo.totalCount)
			actualPred := tc.client.Calls[1].Arguments[1].(*store.SelectionPredicate)

			assert.Equal(t, tc.expectedDescending, actualPred.Descending)
			assert.Equal(t, tc.expectedOrdering, actualPred.Ordering)
			assert.Equal(t, tc.expectedLimit, actualPred.Limit)
			assert.Equal(t, tc.expectedOffset, actualPred.Offset)
		})
	}
}

func TestNamespaceTypeEventFiltersField(t *testing.T) {
	client := new(MockEventFilterClient)
	client.On("ListEventFilters", mock.Anything).Return([]*corev2.EventFilter{
		corev2.FixtureEventFilter("a"),
		corev2.FixtureEventFilter("b"),
		corev2.FixtureEventFilter("c"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceEventFiltersFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	cfg := ServiceConfig{EventFilterClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev3.FixtureNamespace("default")
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
	params := schema.NamespaceHandlersFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	cfg := ServiceConfig{HandlerClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev3.FixtureNamespace("default")
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
	params := schema.NamespaceMutatorsFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	cfg := ServiceConfig{MutatorClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev3.FixtureNamespace("default")
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
	params := schema.NamespaceSilencesFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = corev3.FixtureNamespace("xxx")

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
