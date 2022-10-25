package etcdstore_test

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
	oldstore "github.com/sensu/sensu-go/backend/store/etcd"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func testWithEtcdClient(t *testing.T, f func(storev2.Interface, *clientv3.Client)) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()

	s := etcdstore.NewStore(client)
	oldStore := oldstore.NewStore(client)
	ns := &corev2.Namespace{Name: "default"}

	if err := oldStore.CreateNamespace(context.Background(), ns); err != nil {
		t.Fatal(err)
	}

	f(s, client)
}
