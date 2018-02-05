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
	client, err := e.NewClient()
	require.NoError(t, err)

	queue := New("testenq", client)
	err = queue.Enqueue(context.Background(), "test item")
	assert.NoError(t, err)
}

func TestDequeueSingleItem(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)

	queue := New("testdeq", client)
	err = queue.Enqueue(context.Background(), "test single item dequeue")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test single item dequeue", item.Value)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	cancel()
	item, err = queue.Dequeue(ctx)
	require.Error(t, err)
}

func TestDequeueFIFO(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)

	queue := New("testfifo", client)
	items := []string{"hello", "there", "world", "asdf", "fjdksl", "lalalal"}

	for _, item := range items {
		require.NoError(t, queue.Enqueue(context.Background(), item))
	}

	result := []string{}

	for range items {
		item, err := queue.Dequeue(context.Background())
		require.NoError(t, err)
		result = append(result, item.Value)
	}

	require.Equal(t, items, result)
}

func TestDequeueParallel(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)

	queue := New("testparallel", client)
	items := map[string]struct{}{
		"hello":   struct{}{},
		"there":   struct{}{},
		"world":   struct{}{},
		"asdf":    struct{}{},
		"fjdksl":  struct{}{},
		"lalalal": struct{}{},
	}

	var wg sync.WaitGroup
	wg.Add(len(items))

	for item := range items {
		go func(item string) {
			require.NoError(t, queue.Enqueue(context.Background(), item))
			wg.Done()
		}(item)
	}

	wg.Wait()

	results := make(map[string]struct{})
	var mu sync.Mutex
	wg.Add(len(items))

	for range items {
		go func() {
			item, err := queue.Dequeue(context.Background())
			require.NoError(t, err)
			mu.Lock()
			results[item.Value] = struct{}{}
			mu.Unlock()
			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, items, results)
}

func TestNack(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)

	queue := New("testnack", client)
	err = queue.Enqueue(context.Background(), "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value)

	err = item.Nack(context.Background())
	require.NoError(t, err)

	item, err = queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value)
}

func TestAck(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)

	queue := New("testack", client)
	err = queue.Enqueue(context.Background(), "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value)

	err = item.Ack(context.Background())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	item, err = queue.Dequeue(ctx)
	require.Error(t, err)
}

func TestOnce(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)

	queue := New("testonce", client)

	err = queue.Enqueue(context.Background(), "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value)

	// nack should return the original itemt to the queue for reprocessing, ack
	// called after should have no effect
	require.NoError(t, item.Nack(context.Background()))
	require.NoError(t, item.Ack(context.Background()))
	nackedItem, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, item.Value, nackedItem.Value)
}

func TestNackExpired(t *testing.T) {
	t.Parallel()
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)

	queue := New("testexpired", client)
	queue.itemTimeout = 2 * time.Second

	ctx, cancel := context.WithCancel(context.Background())

	err = queue.Enqueue(ctx, "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(ctx)
	require.NoError(t, err)

	// close the first client
	err = client.Close()
	require.NoError(t, err)
	cancel()

	// create a new client and queue
	newClient, err := e.NewClient()
	require.NoError(t, err)

	// wait to make sure the item has timed out
	time.Sleep(2 * time.Second)

	newQueue := New("testexpired", newClient)
	newQueue.itemTimeout = 2 * time.Second

	// nacked item should go back in the work queue lane
	item, err = newQueue.Dequeue(context.Background())
	require.NoError(t, err)

	require.Equal(t, "test item", item.Value)
}
