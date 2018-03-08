// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testWithEtcd(t *testing.T, f func(store.Store)) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	s, err := NewStore(e)
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "failed to get store from etcd")
	}

	// Mock a default organization & environment
	require.NoError(t, s.CreateOrganization(context.Background(), types.FixtureOrganization("default")))

	f(s)
}
