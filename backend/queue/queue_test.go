// +build !race

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

	queue := New("testenq", client, time.Second)
	err = queue.Enqueue(context.Background(), "test item")
	assert.NoError(t, err)
}

func TestDequeueSingleItem(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)

	queue := New("testdeq", client, time.Second)
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

	queue := New("testfifo", client, time.Second)
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

	queue := New("testparallel", client, time.Second)
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

	queue := New("testnack", client, time.Second)
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

	queue := New("testack", client, time.Second)
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

	queue := New("testonce", client, time.Second)

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

	queue := New("testexpired", client, time.Second)

	err = queue.Enqueue(context.Background(), "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)

	// wait to make sure the item has timed out
	time.Sleep(2 * time.Second)

	// nacked item should go back in the work queue lane
	item, err = queue.Dequeue(context.Background())
	require.NoError(t, err)

	require.Equal(t, "test item", item.Value)
}
