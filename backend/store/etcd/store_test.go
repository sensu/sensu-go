// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/require"
)

func testWithEtcd(t *testing.T, f func(store.Store)) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)

	s := NewStore(client, e.Name())

	// Mock a default namespace
	require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("default")))

	f(s)
}
