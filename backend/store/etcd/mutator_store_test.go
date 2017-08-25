package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestMutatorStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		mutator := types.FixtureMutator("mutator1")
		ctx := context.WithValue(context.Background(), types.OrganizationKey, mutator.Organization)
		ctx = context.WithValue(ctx, types.EnvironmentKey, mutator.Environment)

		// We should receive an empty slice if no results were found
		mutators, err := store.GetMutators(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, mutators)

		err = store.UpdateMutator(ctx, mutator)
		assert.NoError(t, err)

		retrieved, err := store.GetMutatorByName(ctx, "mutator1")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, mutator.Name, retrieved.Name)
		assert.Equal(t, mutator.Command, retrieved.Command)
		assert.Equal(t, mutator.Timeout, retrieved.Timeout)

		mutators, err = store.GetMutators(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, mutators)
		assert.Equal(t, 1, len(mutators))

		// Updating a mutator in a nonexistent org and env should not work
		mutator.Organization = "missing"
		mutator.Environment = "missing"
		err = store.UpdateMutator(ctx, mutator)
		assert.Error(t, err)
	})
}
