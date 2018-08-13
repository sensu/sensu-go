// +build integration,!race

package ring

import (
	"context"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	ring := EtcdGetter{Client: client, BackendID: "TestAdd"}.GetRing("testadd")
	err = ring.Add(context.Background(), "foo")
	assert.NoError(t, err)
}

func TestRemove(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	ring := EtcdGetter{Client: client, BackendID: "TestRemove"}.GetRing("testremove")
	require.NoError(t, ring.Add(context.Background(), "foo"))
	require.NoError(t, ring.Remove(context.Background(), "foo"))
}

func TestNext(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	ring := EtcdGetter{Client: client, BackendID: "TestNext"}.GetRing("testnext")

	items := []string{"foo", "bar", "baz"}
	for _, item := range items {
		require.NoError(t, ring.Add(context.Background(), item))
	}

	var got []string

	for i := 0; i < 9; i++ {
		item, err := ring.Next(context.Background())
		require.NoError(t, err)
		got = append(got, item)
	}

	want := append(items, items...)
	want = append(want, items...)

	assert.Equal(t, want, got)

	require.NoError(t, ring.Remove(context.Background(), "bar"))

	newItems := []string{"foo", "baz"}
	want = want[:0]
	for i := 0; i < 5; i++ {
		want = append(want, newItems...)
	}

	got = got[:0]

	for i := 0; i < 10; i++ {
		item, err := ring.Next(context.Background())
		require.NoError(t, err)
		got = append(got, item)
	}

	assert.Equal(t, want, got)
}

func TestErrorOnNext(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	getterA := EtcdGetter{Client: client, BackendID: "TestErrorOnNextA"}

	getterB := EtcdGetter{Client: client, BackendID: "TestErrorOnNextB"}

	r1 := getterA.GetRing("blocknext")
	r2 := getterB.GetRing("blocknext")

	require.NoError(t, r1.Add(context.Background(), "foo"))
	require.NoError(t, r2.Add(context.Background(), "bar"))

	_, err = r2.Next(context.Background())
	if err != ErrNotOwner {
		t.Fatalf("wanted ErrNotOwner, got %v", err)
	}

	value, err := r1.Next(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "foo", value)

	value, err = r2.Next(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "bar", value)

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
	require.NoError(t, err)
	defer client.Close()

	getter := EtcdGetter{Client: client, BackendID: "TestTransferOwner"}

	r1 := getter.GetRing("testtransfer")
	r2 := getter.GetRing("testtransfer")
	r2.(*Ring).backendID = "something-else-entirely"

	require.NoError(t, r1.Add(context.Background(), "foo"))
	require.NoError(t, r2.Add(context.Background(), "foo"))

	value, err := r2.Next(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "foo", value)

	if _, err := r1.Next(context.Background()); err != ErrNotOwner {
		t.Fatalf("wanted ErrNotOwner, got %v", err)
	}
}

func TestErrNotOwner(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	getter := EtcdGetter{Client: client, BackendID: "TestErrNotOwner"}

	r1 := getter.GetRing("testerrnotowner")
	r2 := getter.GetRing("testerrnotowner")
	r2.(*Ring).backendID = "something-else-entirely"

	require.NoError(t, r1.Add(context.Background(), "foo"))
	assert.Equal(t, ErrNotOwner, r2.Remove(context.Background(), "foo"))
}

func TestExpire(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
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
	assert.Equal(t, ErrEmptyRing, err)
}
