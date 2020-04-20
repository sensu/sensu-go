// +build integration,!race

package etcd

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckConfigStorage(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		check := corev2.FixtureCheckConfig("check1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, check.Namespace)

		pred := &store.SelectionPredicate{}

		// We should receive an empty slice if no results were found
		checks, err := s.GetCheckConfigs(ctx, pred)
		assert.NoError(t, err)
		assert.NotNil(t, checks)
		assert.Empty(t, pred.Continue)

		err = s.UpdateCheckConfig(ctx, check)
		require.NoError(t, err)

		retrieved, err := s.GetCheckConfigByName(ctx, "check1")
		assert.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, check.Name, retrieved.Name)
		assert.Equal(t, check.Interval, retrieved.Interval)
		assert.Equal(t, check.Subscriptions, retrieved.Subscriptions)
		assert.Equal(t, check.Command, retrieved.Command)
		assert.Equal(t, check.Stdin, retrieved.Stdin)

		checks, err = s.GetCheckConfigs(ctx, pred)
		assert.NoError(t, err)
		assert.NotEmpty(t, checks)
		assert.Equal(t, 1, len(checks))
		assert.Empty(t, pred.Continue)

		// Updating a check in a nonexistent org and env should not work
		check.Namespace = "missing"
		err = s.UpdateCheckConfig(ctx, check)
		assert.Error(t, err)
	})
}
