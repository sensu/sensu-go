// +build integration,!race

package ringv2

import (
	"context"
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
)

func TestPool(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	pool := NewPool(client)

	fooRing := pool.Get("foo")

	if got, want := pool.Get("foo"), fooRing; !reflect.DeepEqual(got, want) {
		t.Fatal("rings should be equal")
	}

	if got, want := pool.Get("bar"), fooRing; reflect.DeepEqual(got, want) {
		t.Fatal("rings should not be equal")
	}

	ch := fooRing.Watch(context.Background(), 1)

	if got, want := fooRing.Watch(context.Background(), 1), ch; !reflect.DeepEqual(got, want) {
		t.Fatal("watchers should be equal")
	}
}
