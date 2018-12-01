// +build integration,!race

package ring

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/etcd"
)

func noError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	noError(t, err)
	defer client.Close()

	ring := EtcdGetter{Client: client, BackendID: "TestAdd"}.GetRing("testadd")
	noError(t, ring.Add(context.Background(), "foo"))
}

func TestRemove(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	noError(t, err)
	defer client.Close()

	ring := EtcdGetter{Client: client, BackendID: "TestRemove"}.GetRing("testremove")
	noError(t, ring.Add(context.Background(), "foo"))
	noError(t, ring.Remove(context.Background(), "foo"))
}

func TestNext(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	noError(t, err)
	defer client.Close()

	ring := EtcdGetter{Client: client, BackendID: "TestNext"}.GetRing("testnext")

	items := []string{"foo", "bar", "baz"}
	for _, item := range items {
		noError(t, ring.Add(context.Background(), item))
	}

	var got []string

	for i := 0; i < 9; i++ {
		item, err := ring.Next(context.Background())
		noError(t, err)
		got = append(got, item)
	}

	want := append(items, items...)
	want = append(want, items...)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad values: got %v, want %v", got, want)
	}

	noError(t, ring.Remove(context.Background(), "bar"))

	newItems := []string{"foo", "baz"}
	want = want[:0]
	for i := 0; i < 5; i++ {
		want = append(want, newItems...)
	}

	got = got[:0]

	for i := 0; i < 10; i++ {
		item, err := ring.Next(context.Background())
		noError(t, err)
		got = append(got, item)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad values: got %v, want %v", got, want)
	}
}

func TestErrorOnNext(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	noError(t, err)
	defer client.Close()

	getterA := EtcdGetter{Client: client, BackendID: "TestErrorOnNextA"}

	getterB := EtcdGetter{Client: client, BackendID: "TestErrorOnNextB"}

	r1 := getterA.GetRing("blocknext")
	r2 := getterB.GetRing("blocknext")

	noError(t, r1.Add(context.Background(), "foo"))
	noError(t, r2.Add(context.Background(), "bar"))

	_, err = r2.Next(context.Background())
	if err != ErrNotOwner {
		t.Fatalf("wanted ErrNotOwner, got %v", err)
	}

	value, err := r1.Next(context.Background())
	noError(t, err)
	if got, want := value, "foo"; got != want {
		t.Fatalf("bad values: got %q, want %q", got, want)
	}

	value, err = r2.Next(context.Background())
	noError(t, err)
	if got, want := value, "bar"; got != want {
		t.Fatalf("bad values: got %q, want %q", got, want)
	}

	_, err = r1.Next(context.Background())
	if err != ErrNotOwner {
		t.Fatalf("wanted ErrNotOwner, got %v", err)
	}
}

func TestTransferOwnership(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	noError(t, err)
	defer client.Close()

	getter := EtcdGetter{Client: client, BackendID: "TestTransferOwner"}

	r1 := getter.GetRing("testtransfer")
	r2 := getter.GetRing("testtransfer")
	r2.(*Ring).backendID = "something-else-entirely"

	noError(t, r1.Add(context.Background(), "foo"))
	noError(t, r2.Add(context.Background(), "foo"))

	value, err := r2.Next(context.Background())
	noError(t, err)

	if got, want := value, "foo"; got != want {
		t.Fatalf("bad values: got %q, want %q", got, want)
	}

	if _, err := r1.Next(context.Background()); err != ErrNotOwner {
		t.Fatalf("wanted ErrNotOwner, got %v", err)
	}
}

func TestErrNotOwner(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	noError(t, err)
	defer client.Close()

	getter := EtcdGetter{Client: client, BackendID: "TestErrNotOwner"}

	r1 := getter.GetRing("testerrnotowner")
	r2 := getter.GetRing("testerrnotowner")
	r2.(*Ring).backendID = "something-else-entirely"

	noError(t, r1.Add(context.Background(), "foo"))
	if got, want := r2.Remove(context.Background(), "foo"), ErrNotOwner; got != want {
		t.Fatalf("bad error: got %v, want %v", got, want)
	}
}

func TestExpire(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	noError(t, err)
	defer client.Close()

	ring := EtcdGetter{Client: client, BackendID: "TestExpire"}.GetRing("testexpire").(*Ring)
	ring.leaseTimeout = 1

	if err := ring.Add(context.Background(), "foo"); err != nil {
		t.Fatal(err)
	}

	// Simulate the client dying
	if _, err := ring.client.Revoke(context.Background(), leaseIDCache[ring.Name]); err != nil {
		t.Fatal(err)
	}

	// Give the cluster some time to expire the lease. Unfortunately there
	// doesn't seem to be any way to be informed of when the lease expires.
	time.Sleep(time.Second * 5)

	_, err = ring.Next(context.Background())
	if got, want := err, ErrEmptyRing; got != want {
		t.Fatalf("bad error: got %v, want %v", got, want)
	}
}
