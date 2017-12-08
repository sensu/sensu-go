package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, types.OrganizationKey, "default")
		ctx = context.WithValue(ctx, types.EnvironmentKey, "foo")

		org := "default"
		env := types.FixtureEnvironment("foo")
		err := store.UpdateEnvironment(ctx, env)
		assert.NoError(t, err)

		result, err := store.GetEnvironment(ctx, env.Organization, env.Name)
		assert.NoError(t, err)
		assert.Equal(t, env.Name, result.Name)

		// Missing environment returns nil
		_, err = store.GetEnvironment(ctx, org, "missing")
		assert.NoError(t, err)

		// Get all environments
		envs, err := store.GetEnvironments(ctx, org)
		assert.NoError(t, err)
		assert.NotEmpty(t, envs)
		assert.Equal(t, 2, len(envs))

		// Delete a non-empty environment
		exCheck := types.FixtureCheckConfig("id")
		exCheck.Environment = env.Name
		exCheck.Organization = org
		require.NoError(t, store.UpdateCheckConfig(ctx, exCheck))
		err = store.DeleteEnvironment(ctx, env)
		assert.Error(t, err)

		// Delete a non-empty environment w/ role
		require.NoError(t, store.DeleteCheckConfigByName(ctx, exCheck.Name))
		require.NoError(t, store.UpdateRole(ctx, types.FixtureRole("1", org, env.Name)))
		err = store.DeleteEnvironment(ctx, env)
		assert.Error(t, err)

		// Delete an environment
		require.NoError(t, store.DeleteRoleByName(ctx, "1"))
		err = store.DeleteEnvironment(ctx, env)
		assert.NoError(t, err)

		// Delete a missing org
		err = store.DeleteEnvironment(ctx, &types.Environment{Organization: org, Name: "missing"})
		assert.Error(t, err)

		// Create a environment within a missing org
		env.Organization = "missing"
		err = store.UpdateEnvironment(ctx, env)
		assert.Error(t, err)

		// Retrieve all environments again. We should have the default one
		envs, err = store.GetEnvironments(ctx, org)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(envs))
	})
}
