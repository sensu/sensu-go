package graphql

import (
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryTypeEventField(t *testing.T) {
	mock := mockEventFetcher{&types.Event{}, nil}
	impl := queryImpl{eventFinder: mock}

	args := schema.QueryEventFieldResolverArgs{Ns: schema.NewNamespaceInput("a", "b")}
	params := schema.QueryEventFieldResolverParams{Args: args}

	res, err := impl.Event(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestQueryTypeEnvironmentField(t *testing.T) {
	mock := mockEnvironmentFinder{&types.Environment{}, nil}
	impl := queryImpl{envFinder: mock}

	params := schema.QueryEnvironmentFieldResolverParams{}
	params.Args.Environment = "us-west-2"
	params.Args.Organization = "bobs-burgers"

	res, err := impl.Environment(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestQueryTypeEntityField(t *testing.T) {
	mock := mockEntityFetcher{types.FixtureEntity("abc"), nil}
	impl := queryImpl{entityFinder: mock}

	params := schema.QueryEntityFieldResolverParams{}
	params.Args.Ns = schema.NewNamespaceInput("org", "env")
	params.Args.Name = "abc"

	res, err := impl.Entity(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}
