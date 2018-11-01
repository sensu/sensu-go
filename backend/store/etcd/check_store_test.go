// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckConfigStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		check := types.FixtureCheckConfig("check1")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, check.Namespace)

		// We should receive an empty slice if no results were found
		checks, err := store.GetCheckConfigs(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, checks)

		err = store.UpdateCheckConfig(ctx, check)
		require.NoError(t, err)

		retrieved, err := store.GetCheckConfigByName(ctx, "check1")
		assert.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, check.Name, retrieved.Name)
		assert.Equal(t, check.Interval, retrieved.Interval)
		assert.Equal(t, check.Subscriptions, retrieved.Subscriptions)
		assert.Equal(t, check.Command, retrieved.Command)
		assert.Equal(t, check.Stdin, retrieved.Stdin)

		checks, err = store.GetCheckConfigs(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, checks)
		assert.Equal(t, 1, len(checks))

		// Updating a check in a nonexistent org and env should not work
		check.Namespace = "missing"
		err = store.UpdateCheckConfig(ctx, check)
		assert.Error(t, err)
	})
}
