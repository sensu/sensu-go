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
