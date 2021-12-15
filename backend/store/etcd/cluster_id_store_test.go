//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/backend/store"
)

func TestClusterIDStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		ctx := context.Background()

		// The first call to GetClusterID creates a new cluster ID
		id, err := store.GetClusterID(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// The id is a valid UUID
		if _, err := uuid.Parse(id); err != nil {
			t.Fatal(err)
		}

		id2, err := store.GetClusterID(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if got, want := id2, id; got != want {
			t.Fatal("cluster IDs differ")
		}
	})
}
