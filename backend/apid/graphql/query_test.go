package graphql

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockQueryEventFetcher struct {
	record *types.Event
	err    error
}

func (m mockQueryEventFetcher) Find(ctx context.Context, entity, check string) (*types.Event, error) {
	return m.record, m.err
}

type mockQueryEnvironmentFetcher struct {
	record *types.Environment
	err    error
}

func (m mockQueryEnvironmentFetcher) Find(ctx context.Context, org, env string) (*types.Environment, error) {
	if org != "bobs-burgers" || env != "us-west-2" {
		return nil, nil
	}
	return m.record, m.err
}

func TestQueryTypeEventField(t *testing.T) {
	mock := mockQueryEventFetcher{&types.Event{}, nil}
	impl := queryImpl{eventCtrl: mock}

	args := schema.QueryEventFieldResolverArgs{Ns: schema.NewNamespaceInput("a", "b")}
	params := schema.QueryEventFieldResolverParams{Args: args}

	res, err := impl.Event(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestQueryTypeEnvironmentField(t *testing.T) {
	mock := mockQueryEnvironmentFetcher{&types.Environment{}, nil}
	impl := queryImpl{environmentCtrl: mock}

	params := schema.QueryEnvironmentFieldResolverParams{}
	params.Args.Environment = "us-west-2"
	params.Args.Organization = "bobs-burgers"

	res, err := impl.Environment(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}
