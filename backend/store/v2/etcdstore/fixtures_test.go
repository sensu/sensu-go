package etcdstore_test

import (
	"errors"
	"io/ioutil"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-api-tools/apis"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	etcdstorev2 "github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sirupsen/logrus"
)

type testResource struct {
	Metadata *corev2.ObjectMeta
}

func (t *testResource) GetMetadata() *corev2.ObjectMeta {
	return t.Metadata
}

func (t *testResource) SetMetadata(m *corev2.ObjectMeta) {
	t.Metadata = m
}

func (t *testResource) StoreName() string {
	return "testresource"
}

func (t *testResource) RBACName() string {
	return "testresource"
}

func (t *testResource) URIPath() string {
	return "api/backend/store/namespaces/default/testresource/test"
}

func (t *testResource) Validate() error {
	return nil
}

func (t *testResource) GetTypeMeta() corev2.TypeMeta {
	return corev2.TypeMeta{
		Type:       "testResource",
		APIVersion: "store/wrap_test",
	}
}

func fixtureResolver(name string) (interface{}, error) {
	switch name {
	case "testResource":
		return &testResource{}, nil
	default:
		return nil, errors.New("type does not exist")
	}
}

func init() {
	apis.RegisterResolver("store/wrap_test", fixtureResolver)
}

func testWithEtcdStore(t testing.TB, f func(*etcdstorev2.Store)) {
	logrus.SetOutput(ioutil.Discard)
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	s := etcdstorev2.NewStore(client)
	f(s)
}

func testWithV1Store(t testing.TB, f func(store.Store)) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	s := etcdstore.NewStore(client, "default")

	f(s)
}
