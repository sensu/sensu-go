// +build integration,!race

package ringv2

import (
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
)

func TestPool(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	pool := NewPool(client)

	fooRing := pool.Get("foo")

	if got, want := pool.Get("foo"), fooRing; !reflect.DeepEqual(got, want) {
		t.Fatal("rings should be equal")
	}

	if got, want := pool.Get("bar"), fooRing; reflect.DeepEqual(got, want) {
		t.Fatal("rings should not be equal")
	}
}
