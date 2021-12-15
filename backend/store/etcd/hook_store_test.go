//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestHookConfigStorage(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		hook := corev2.FixtureHookConfig("hook1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, hook.Namespace)

		// We should receive an empty slice if no results were found
		pred := &store.SelectionPredicate{}
		hooks, err := s.GetHookConfigs(ctx, pred)
		assert.NoError(t, err)
		assert.NotNil(t, hooks)
		assert.Empty(t, pred.Continue)

		err = s.UpdateHookConfig(ctx, hook)
		require.NoError(t, err)

		retrieved, err := s.GetHookConfigByName(ctx, "hook1")
		assert.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, hook.Name, retrieved.Name)
		assert.Equal(t, hook.Command, retrieved.Command)
		assert.Equal(t, hook.Timeout, retrieved.Timeout)
		assert.Equal(t, hook.Stdin, retrieved.Stdin)

		hooks, err = s.GetHookConfigs(ctx, pred)
		assert.NoError(t, err)
		assert.NotEmpty(t, hooks)
		assert.Equal(t, 1, len(hooks))
		assert.Empty(t, pred.Continue)

		// Updating a hook in a nonexistent org and env should not work
		hook.Namespace = "missing"
		err = s.UpdateHookConfig(ctx, hook)
		assert.Error(t, err)
	})
}
