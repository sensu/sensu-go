// +build integration,!race

package queue

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnqueue(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client := e.NewEmbeddedClient()
	defer client.Close()

	backendID := etcd.NewBackendIDGetter(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := backendID.Start(ctx); err != nil {
		t.Fatal(err)
	}

	queue := New("testenq", client, backendID)
	err := queue.Enqueue(context.Background(), "test item")
	assert.NoError(t, err)
}

func TestDequeueSingleItem(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client := e.NewEmbeddedClient()
	defer client.Close()

	backendID := etcd.NewBackendIDGetter(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := backendID.Start(ctx); err != nil {
		t.Fatal(err)
	}
	queue := New("testdeq", client, backendID)
	err := queue.Enqueue(context.Background(), "test single item dequeue")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test single item dequeue", item.Value())

	defer func() {
		assert.NoError(t, item.Ack(context.Background()))
	}()

	ctx, cancel = context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, err = queue.Dequeue(ctx)
	assert.Error(t, err)
}

func TestDequeueFIFO(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client := e.NewEmbeddedClient()
	defer client.Close()

	backendID := etcd.NewBackendIDGetter(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := backendID.Start(ctx); err != nil {
		t.Fatal(err)
	}
	queue := New("testfifo", client, backendID)
	items := []string{"hello", "there", "world", "asdf", "fjdksl", "lalalal"}

	for _, item := range items {
		require.NoError(t, queue.Enqueue(context.Background(), item))
	}

	result := []string{}

	for range items {
		item, err := queue.Dequeue(context.Background())
		require.NoError(t, err)
		result = append(result, item.Value())
	}

	require.Equal(t, items, result)
}

func TestNack(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client := e.NewEmbeddedClient()
	defer client.Close()

	backendID := etcd.NewBackendIDGetter(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := backendID.Start(ctx); err != nil {
		t.Fatal(err)
	}
	queue := New("testnack", client, backendID)
	err := queue.Enqueue(context.Background(), "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value())

	err = item.Nack(context.Background())
	require.NoError(t, err)

	item, err = queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value())
}

func TestAck(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client := e.NewEmbeddedClient()
	defer client.Close()

	backendID := etcd.NewBackendIDGetter(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := backendID.Start(ctx); err != nil {
		t.Fatal(err)
	}
	queue := New("testack", client, backendID)
	err := queue.Enqueue(context.Background(), "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value())

	err = item.Ack(context.Background())
	require.NoError(t, err)

	ctx, cancel = context.WithTimeout(ctx, time.Second)
	defer cancel()
	_, err = queue.Dequeue(ctx)
	require.Error(t, err)
}

func TestOnce(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client := e.NewEmbeddedClient()
	defer client.Close()

	backendID := etcd.NewBackendIDGetter(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := backendID.Start(ctx); err != nil {
		t.Fatal(err)
	}
	queue := New("testonce", client, backendID)

	err := queue.Enqueue(context.Background(), "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value())

	// nack should return the original itemt to the queue for reprocessing, ack
	// called after should have no effect
	require.NoError(t, item.Nack(context.Background()))
	require.NoError(t, item.Ack(context.Background()))
	nackedItem, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, item.Value(), nackedItem.Value())
}

func TestMultipleSubscribers(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	backendID1 := etcd.NewBackendIDGetter(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := backendID1.Start(ctx); err != nil {
		t.Fatal(err)
	}
	backendID2 := etcd.NewBackendIDGetter(client)
	if err := backendID2.Start(ctx); err != nil {
		t.Fatal(err)
	}
	backendID3 := etcd.NewBackendIDGetter(client)
	if err := backendID3.Start(ctx); err != nil {
		t.Fatal(err)
	}

	// Each queue is associated with a different backend
	q1 := New("testMultipleSubscribers", client, backendID1)
	q2 := New("testMultipleSubscribers", client, backendID2)
	q3 := New("testMultipleSubscribers", client, backendID3)

	ctx, cancel = context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	require.NoError(t, q1.Enqueue(ctx, "foobar"))

	// Each queue gets the message!
	for _, q := range []*Queue{q1, q2, q3} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		item, err := q.Dequeue(ctx)
		require.NoError(t, err)
		assert.Equal(t, "foobar", item.Value())
	}
}

func TestDequeueParallel(t *testing.T) {
	t.Parallel()
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client := e.NewEmbeddedClient()
	defer client.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	backendID := etcd.NewBackendIDGetter(client)
	if err := backendID.Start(ctx); err != nil {
		t.Fatal(err)
	}
	queue := New("testparallel", client, backendID)
	items := map[string]struct{}{
		"hello":   struct{}{},
		"there":   struct{}{},
		"world":   struct{}{},
		"asdf":    struct{}{},
		"fjdksl":  struct{}{},
		"lalalal": struct{}{},
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(items))
	var errEnqueue error
	for item := range items {
		go func(item string) {
			defer wg.Done()
			// Prevent data races when inspecting the error
			mu.Lock()
			defer mu.Unlock()
			if err := queue.Enqueue(context.Background(), item); err != nil {
				errEnqueue = err
			}
		}(item)
	}
	wg.Wait()
	// Make sure we didn't encounter any error when adding items to the queue.
	// If we had multiple errors, only the last one is saved
	require.NoError(t, errEnqueue)
	results := make(map[string]struct{})
	wg.Add(len(items))
	for range items {
		go func() {
			defer wg.Done()
			item, err := queue.Dequeue(context.Background())
			// Prevent data races when inspecting the error or the result
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				return
			}
			results[item.Value()] = struct{}{}
		}()
	}
	wg.Wait()
	// Make sure we didn't encounter any error while dequeuing items from the
	// queue. If we had multiple errors, only the last one is saved
	require.NoError(t, errEnqueue)
	assert.Equal(t, items, results)
}
