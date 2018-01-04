package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/store/etcd"
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
	err = queue.Enqueue(context.Background(), "test value")
	assert.NoError(t, err)
}

func TestDequeue(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)

	queue := New("testdeq", client)
	err = queue.Enqueue(context.Background(), "test value")
	require.NoError(t, err)

	value, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test value", value)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = queue.Dequeue(ctx)
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
		value, err := queue.Dequeue(context.Background())
		require.NoError(t, err)
		result = append(result, value)
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
			value, err := queue.Dequeue(context.Background())
			require.NoError(t, err)
			fmt.Printf("value: %q\n", value)
			mu.Lock()
			results[value] = struct{}{}
			mu.Unlock()
			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, items, results)
}
