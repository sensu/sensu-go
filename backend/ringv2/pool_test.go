//go:build integration && !race
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

func TestPoolSetNewFunc(t *testing.T) {
	pool := NewRingPool(func(path string) Interface {
		return new(Ring)
	})
	fooRing := pool.Get("foo")

	pool.SetNewFunc(func(path string) Interface {
		return nil
	})

	fooRing2 := pool.Get("foo")

	if fooRing == fooRing2 {
		t.Fatal("rings should differ")
	}
}
