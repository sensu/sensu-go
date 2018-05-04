package graphql

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

type mockEntityFetcher struct {
	record *types.Entity
	err    error
}

func (m mockEntityFetcher) Find(_ context.Context, _ string) (*types.Entity, error) {
	return m.record, m.err
}

type mockEventQuerier struct {
	els []*types.Event
	err error
}

func (f mockEventQuerier) Query(ctx context.Context, entity, check string) ([]*types.Event, error) {
	return f.els, f.err
}
