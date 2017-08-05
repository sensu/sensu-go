package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestEnvStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		ctx := context.Background()

		org := "default"
		env := types.FixtureEnvironment("foo")
		err := store.UpdateEnvironment(ctx, org, env)
		assert.NoError(t, err)

		result, err := store.GetEnvironment(ctx, org, env.Name)
		assert.NoError(t, err)
		assert.Equal(t, env.Name, result.Name)

		// Missing environment
		_, err = store.GetEnvironment(ctx, org, "missing")
		assert.Error(t, err)

		// Get all environments
		envs, err := store.GetEnvironments(ctx, org)
		assert.NoError(t, err)
		assert.NotEmpty(t, envs)
		assert.Equal(t, 1, len(envs))

		// Delete an environment
		err = store.DeleteEnvironment(ctx, org, env.Name)
		assert.NoError(t, err)

		// Delete a missing org
		err = store.DeleteEnvironment(ctx, org, "missing")
		assert.Error(t, err)

		// Create a environment within a missing org
		err = store.UpdateEnvironment(ctx, "missing", env)
		assert.Error(t, err)

		// Retrieve all environments again
		envs, err = store.GetEnvironments(ctx, org)
		assert.NoError(t, err)
		assert.Empty(t, envs)
	})
}
