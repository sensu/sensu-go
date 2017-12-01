package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHookConfigStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		hook := types.FixtureHookConfig("hook1")
		ctx := context.WithValue(context.Background(), types.OrganizationKey, hook.Organization)
		ctx = context.WithValue(ctx, types.EnvironmentKey, hook.Environment)

		// We should receive an empty slice if no results were found
		hooks, err := store.GetHookConfigs(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, hooks)

		err = store.UpdateHookConfig(ctx, hook)
		require.NoError(t, err)

		retrieved, err := store.GetHookConfigByName(ctx, "hook1")
		assert.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, hook.Name, retrieved.Name)
		assert.Equal(t, hook.Command, retrieved.Command)
		assert.Equal(t, hook.Timeout, retrieved.Timeout)
		assert.Equal(t, hook.Stdin, retrieved.Stdin)

		hooks, err = store.GetHookConfigs(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, hooks)
		assert.Equal(t, 1, len(hooks))

		// Updating a hook in a nonexistent org and env should not work
		hook.Organization = "missing"
		hook.Environment = "missing"
		err = store.UpdateHookConfig(ctx, hook)
		assert.Error(t, err)
	})
}
