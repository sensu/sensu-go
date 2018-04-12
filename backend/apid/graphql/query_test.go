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

func TestQueryTypeEventField(t *testing.T) {
	mock := mockQueryEventFetcher{&types.Event{}, nil}
	impl := queryImpl{eventController: mock}

	args := schema.QueryEventFieldResolverArgs{Ns: schema.NewNamespaceInput("a", "b")}
	params := schema.QueryEventFieldResolverParams{Args: args}

	res, err := impl.Event(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}
