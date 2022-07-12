package etcdstore_test

import (
	"context"
	"testing"

	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestGetSetDatabaseVersion(t *testing.T) {
	testWithEtcdClient(t, func(_ storev2.Interface, client *clientv3.Client) {
		ctx := context.Background()
		version, err := etcdstore.GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 0; got != want {
			t.Errorf("bad db version: got %d, want %d", got, want)
		}
		if err := etcdstore.SetDatabaseVersion(ctx, client, 3); err != nil {
			t.Fatal(err)
		}
		version, err = etcdstore.GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 3; got != want {
			t.Errorf("bad db version: got %d, want %d", got, want)
		}
	})
}
