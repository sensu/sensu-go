package etcdstore_test

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	oldstore "github.com/sensu/sensu-go/backend/store/etcd"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	entityKeyBuilder       = store.NewKeyBuilder((&corev2.Entity{}).StoreName())
	entityConfigKeyBuilder = store.NewKeyBuilder((&corev3.EntityConfig{}).StoreName())
)

func getEntityPath(entity *corev2.Entity) string {
	return entityKeyBuilder.WithResource(entity).Build(entity.Name)
}

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
