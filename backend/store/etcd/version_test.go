// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"go.etcd.io/etcd/client/v3"
)

func TestGetSetDatabaseVersion(t *testing.T) {
	testWithEtcdClient(t, func(_ store.Store, client *clientv3.Client) {
		ctx := context.Background()
		version, err := GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 0; got != want {
			t.Errorf("bad db version: got %d, want %d", got, want)
		}
		if err := SetDatabaseVersion(ctx, client, 3); err != nil {
			t.Fatal(err)
		}
		version, err = GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 3; got != want {
			t.Errorf("bad db version: got %d, want %d", got, want)
		}
	})
}
