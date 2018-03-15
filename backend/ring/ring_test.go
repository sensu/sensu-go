// +build integration,!race

package ring

import (
	"context"
	"sync"
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

	ring := EtcdGetter{client}.GetRing("testadd")
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

	ring := EtcdGetter{client}.GetRing("testremove")
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

	ring := EtcdGetter{client}.GetRing("testnext")

	items := []string{"foo", "bar", "baz"}
	for _, item := range items {
		require.NoError(t, ring.Add(context.Background(), item))
	}

	for i := 0; i < 9; i++ {
		item, err := ring.Next(context.Background())
		require.NoError(t, err)
		assert.Equal(t, items[i%3], item)
	}

	require.NoError(t, ring.Remove(context.Background(), "bar"))

	newItems := []string{"foo", "baz"}

	for i := 0; i < 10; i++ {
		item, err := ring.Next(context.Background())
		require.NoError(t, err)
		assert.Equal(t, newItems[i%2], item)
	}
}

func TestPeek(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	ring := EtcdGetter{client}.GetRing("testpeek")

	items := []string{"foo", "bar", "baz"}
	for _, item := range items {
		require.NoError(t, ring.Add(context.Background(), item))
	}

	item, err := ring.Peek(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "foo", item)

	ring = EtcdGetter{client}.GetRing("testempty")
	_, err = ring.Peek(context.Background())
	assert.Equal(t, ErrEmptyRing, err)
}

func TestBlockOnNext(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	getter := EtcdGetter{client}

	r1 := getter.GetRing("blocknext")
	r2 := getter.GetRing("blocknext")
	r2.(*Ring).backendID = "something-else-entirely"

	require.NoError(t, r1.Add(context.Background(), "foo"))
	require.NoError(t, r2.Add(context.Background(), "bar"))

	var wg sync.WaitGroup
	var errNext error
	wg.Add(1)

	go func() {
		defer wg.Done()
		value, err := r2.Next(context.Background())
		if err != nil {
			errNext = err
			return
		}

		assert.Equal(t, "bar", value)
	}()

	value, err := r1.Next(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "foo", value)

	wg.Wait()

	// Make sure we didn't encountered any error while getting the next item
	require.NoError(t, errNext)

	value, err = r1.Next(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "foo", value)
}

func TestTransferOwnership(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	getter := EtcdGetter{client}

	r1 := getter.GetRing("testtransfer")
	r2 := getter.GetRing("testtransfer")
	r2.(*Ring).backendID = "something-else-entirely"

	require.NoError(t, r1.Add(context.Background(), "foo"))
	require.NoError(t, r2.Add(context.Background(), "foo"))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-time.After(time.Second)
		cancel()
	}()
	_, err = r1.Next(ctx)
	assert.Equal(t, ctx.Err(), err) // it timed out

	value, err := r2.Next(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "foo", value)
}

func TestErrNotOwner(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	getter := EtcdGetter{client}

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

	ring := EtcdGetter{client}.GetRing("testexpire").(*Ring)
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
