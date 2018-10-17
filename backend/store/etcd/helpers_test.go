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

func TestQuery(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		etcd := store.(*Store)

		// Create a namespace
		require.NoError(t, store.CreateNamespace(context.Background(), types.FixtureNamespace("acme")))

		// Create /checks/default/check1
		check1 := types.FixtureCheckConfig("check1")
		// ctx := context.WithValue(context.Background(), types.NamespaceKey, check1.Namespace)
		if err := store.UpdateCheckConfig(context.Background(), check1); err != nil {
			assert.FailNow(t, err.Error())
		}

		// Create /checks/acme/check2
		check2 := types.FixtureCheckConfig("check2")
		check2.Namespace = "acme"
		ctx := context.WithValue(context.Background(), types.NamespaceKey, check2.Namespace)
		if err := store.UpdateCheckConfig(ctx, check2); err != nil {
			assert.FailNow(t, err.Error())
		}

		// Mock a context that put ourselves in the default namespace
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "default")

		// We only have a single result given our current org & env
		resp, err := query(ctx, etcd, getCheckConfigsPath)
		assert.NoError(t, err)
		assert.Len(t, resp.Kvs, 1)

		// Mock a context to query across all namespaces
		ctx = context.WithValue(ctx, types.NamespaceKey, types.NamespaceTypeAll)

		// We now have two result given our "wildcard" org
		resp, err = query(ctx, etcd, getCheckConfigsPath)
		assert.NoError(t, err)
		assert.Len(t, resp.Kvs, 2)
	})
}
