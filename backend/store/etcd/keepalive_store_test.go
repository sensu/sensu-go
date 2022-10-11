//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestKeepaliveStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		entity := corev2.FixtureEntity("entity")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, entity.Namespace)

		err := store.UpdateFailingKeepalive(ctx, entity, 1)
		assert.NoError(t, err)

		records, err := store.GetFailingKeepalives(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(records))

		// Updating a keepalive in a nonexistent org and env should not work
		entity.Namespace = "missing"
		err = store.UpdateFailingKeepalive(ctx, entity, 1)
		assert.Error(t, err)
	})
}
