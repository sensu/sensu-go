package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestFilterStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		filter := types.FixtureFilter("filter1")
		ctx := context.WithValue(context.Background(), types.OrganizationKey, filter.Organization)
		ctx = context.WithValue(ctx, types.EnvironmentKey, filter.Environment)

		// We should receive an empty slice if no results were found
		filters, err := store.GetFilters(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, filters)

		err = store.UpdateFilter(ctx, filter)
		assert.NoError(t, err)

		retrieved, err := store.GetFilterByName(ctx, "filter1")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, filter.Name, retrieved.Name)
		assert.Equal(t, filter.Action, retrieved.Action)
		assert.Equal(t, filter.Attributes, retrieved.Attributes)

		filters, err = store.GetFilters(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, filters)
		assert.Equal(t, 1, len(filters))

		// Updating a filter in a nonexistent org and env should not work
		filter.Organization = "missing"
		filter.Environment = "missing"
		err = store.UpdateFilter(ctx, filter)
		assert.Error(t, err)
	})
}
