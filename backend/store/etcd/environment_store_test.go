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
		ctx = context.WithValue(ctx, types.OrganizationKey, "default")
		ctx = context.WithValue(ctx, types.EnvironmentKey, "foo")

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
		assert.Equal(t, 2, len(envs))

		// Delete a non-empty environment
		exCheck := types.FixtureCheckConfig("id")
		exCheck.Environment = env.Name
		exCheck.Organization = org
		store.UpdateCheckConfig(ctx, exCheck)
		err = store.DeleteEnvironment(ctx, org, env.Name)
		assert.Error(t, err)

		// Delete a non-empty environment w/ role
		store.DeleteCheckConfigByName(ctx, exCheck.Name)
		store.UpdateRole(types.FixtureRole("1", org, env.Name))
		err = store.DeleteEnvironment(ctx, org, env.Name)
		assert.Error(t, err)

		// Delete an environment
		store.DeleteRoleByName("1")
		err = store.DeleteEnvironment(ctx, org, env.Name)
		assert.NoError(t, err)

		// Delete a missing org
		err = store.DeleteEnvironment(ctx, org, "missing")
		assert.Error(t, err)

		// Create a environment within a missing org
		err = store.UpdateEnvironment(ctx, "missing", env)
		assert.Error(t, err)

		// Retrieve all environments again. We should have the default one
		envs, err = store.GetEnvironments(ctx, org)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(envs))
	})
}
