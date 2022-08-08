package graphql

import (
	"context"
	"errors"
	"reflect"
	"testing"

	dto "github.com/prometheus/client_model/go"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestQueryTypeEventField(t *testing.T) {
	client := new(MockEventClient)
	cfg := ServiceConfig{EventClient: client}
	impl := queryImpl{svc: cfg}

	event := corev2.FixtureEvent("a", "b")
	args := schema.QueryEventFieldResolverArgs{Namespace: "ns", Entity: "a", Check: "b"}
	params := schema.QueryEventFieldResolverParams{Args: args, ResolveParams: graphql.ResolveParams{Context: context.Background()}}

	// Success
	client.On("FetchEvent", mock.Anything, event.Entity.Name, event.Check.Name).Return(event, nil).Once()
	res, err := impl.Event(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestQueryTypeEventFilterField(t *testing.T) {
	client := new(MockEventFilterClient)
	cfg := ServiceConfig{EventFilterClient: client}
	impl := queryImpl{svc: cfg}

	filter := corev2.FixtureEventFilter("a")
	args := schema.QueryEventFilterFieldResolverArgs{Namespace: "ns", Name: "a"}
	params := schema.QueryEventFilterFieldResolverParams{Args: args, ResolveParams: graphql.ResolveParams{Context: context.Background()}}

	// Success
	client.On("FetchEventFilter", mock.Anything, filter.Name).Return(filter, nil).Once()
	res, err := impl.EventFilter(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestQueryTypeNamespaceField(t *testing.T) {
	client := new(MockNamespaceClient)
	cfg := ServiceConfig{NamespaceClient: client}
	impl := queryImpl{svc: cfg}

	nsp := corev3.FixtureNamespace("sensu")
	params := schema.QueryNamespaceFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	params.Args.Name = nsp.Metadata.Name

	// Success
	client.On("FetchNamespace", mock.Anything, nsp.Metadata.Name).Return(nsp, nil).Once()
	res, err := impl.Namespace(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestQueryTypeEntityField(t *testing.T) {
	client := new(MockEntityClient)
	cfg := ServiceConfig{EntityClient: client}
	impl := queryImpl{svc: cfg}

	entity := corev2.FixtureEntity("a")
	params := schema.QueryEntityFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	params.Args.Namespace = entity.Namespace
	params.Args.Name = entity.Name

	// Sucess
	client.On("FetchEntity", mock.Anything, entity.Name).Return(entity, nil).Once()
	res, err := impl.Entity(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestQueryTypeCheckField(t *testing.T) {
	client := new(MockCheckClient)
	cfg := ServiceConfig{CheckClient: client}
	impl := queryImpl{svc: cfg}

	check := corev2.FixtureCheckConfig("a")
	params := schema.QueryCheckFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	params.Args.Namespace = check.Namespace
	params.Args.Name = check.Name

	// Sucess
	client.On("FetchCheck", mock.Anything, check.Name).Return(check, nil).Once()
	res, err := impl.Check(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestQueryTypeHandlerField(t *testing.T) {
	client := new(MockHandlerClient)
	cfg := ServiceConfig{HandlerClient: client}
	impl := queryImpl{svc: cfg}

	handler := corev2.FixtureHandler("a")
	params := schema.QueryHandlerFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	params.Args.Namespace = handler.Namespace
	params.Args.Name = handler.Name

	// Success
	client.On("FetchHandler", mock.Anything, handler.Name).Return(handler, nil).Once()
	res, err := impl.Handler(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestQueryTypeMuatorField(t *testing.T) {
	client := new(MockMutatorClient)
	cfg := ServiceConfig{MutatorClient: client}
	impl := queryImpl{svc: cfg}

	mutator := corev2.FixtureMutator("a")
	params := schema.QueryMutatorFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	params.Args.Namespace = mutator.Namespace
	params.Args.Name = mutator.Name

	// Success
	client.On("FetchMutator", mock.Anything, mutator.Name).Return(mutator, nil).Once()
	res, err := impl.Mutator(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestQueryTypeSuggestField(t *testing.T) {
	client := new(MockGenericClient)
	cfg := ServiceConfig{GenericClient: client}
	impl := queryImpl{svc: cfg}

	prevGlobalFilters := GlobalFilters
	GlobalFilters = CheckFilters()
	defer func() {
		GlobalFilters = prevGlobalFilters
	}()

	params := schema.QuerySuggestFieldResolverParams{ResolveParams: graphql.ResolveParams{Context: context.Background()}}
	params.Args.Namespace = "default"
	params.Args.Ref = "core/v2/check_config/subscriptions"
	params.Args.Filters = []string{"published: true"}
	params.Args.Q = "sql"

	// Success
	client.On("List", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*[]*corev2.CheckConfig)
		*arg = []*corev2.CheckConfig{
			{Publish: true, Subscriptions: []string{"bsd", "psql"}},
			{Publish: false, Subscriptions: []string{"windows", "mssql"}},
		}
	}).Return(nil).Once()
	client.On("SetTypeMeta", mock.Anything).Return(nil)
	res, err := impl.Suggest(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func Test_queryImpl_Metrics(t *testing.T) {
	nameA := "sensu"
	nameB := "etcd"
	tests := []struct {
		name       string
		gatherResp []*dto.MetricFamily
		gatherErr  error
		params     schema.QueryMetricsFieldResolverParams
		want       interface{}
		wantErr    bool
	}{
		{
			name:       "err w/ no resp",
			gatherResp: []*dto.MetricFamily{},
			gatherErr:  errors.New("testing"),
			want:       []interface{}{},
			wantErr:    true,
		},
		{
			name:       "err w/ resp",
			gatherResp: []*dto.MetricFamily{&dto.MetricFamily{}},
			gatherErr:  errors.New("testing"),
			want:       []*dto.MetricFamily{&dto.MetricFamily{}},
			wantErr:    false,
		},
		{
			name:       "returns family",
			gatherResp: []*dto.MetricFamily{&dto.MetricFamily{}},
			gatherErr:  nil,
			want:       []*dto.MetricFamily{&dto.MetricFamily{}},
			wantErr:    false,
		},
		{
			name: "filters by name",
			params: schema.QueryMetricsFieldResolverParams{
				Args: schema.QueryMetricsFieldResolverArgs{
					Name: []string{nameA},
				},
			},
			gatherResp: []*dto.MetricFamily{
				&dto.MetricFamily{Name: &nameA},
				&dto.MetricFamily{Name: &nameB},
			},
			want: []*dto.MetricFamily{
				&dto.MetricFamily{Name: &nameA},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := MockMetricGatherer{}
			mock.On("Gather").Return(tt.gatherResp, tt.gatherErr)
			r := &queryImpl{
				svc: ServiceConfig{
					MetricGatherer: &mock,
				},
			}
			got, err := r.Metrics(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("queryImpl.Metrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queryImpl.Metrics() = %v, want %v", got, tt.want)
			}
		})
	}
}
