package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEtcd(t *testing.T) (*clientv3.Client, func()) {
	tmpDir, remove := testutil.TempDir(t)

	ports := make([]int, 2)
	err := testutil.RandomPorts(ports)
	if err != nil {
		t.Fatal(err)
	}
	clURL := fmt.Sprintf("http://127.0.0.1:%d", ports[0])
	apURL := fmt.Sprintf("http://127.0.0.1:%d", ports[1])
	initCluster := fmt.Sprintf("default=%s", apURL)

	cfg := etcd.NewConfig()
	cfg.DataDir = tmpDir
	cfg.ListenClientURL = clURL
	cfg.ListenPeerURL = apURL
	cfg.InitialCluster = initCluster
	cfg.InitialClusterState = etcd.ClusterStateNew
	cfg.InitialAdvertisePeerURL = apURL
	cfg.Name = "default"

	e, err := etcd.NewEtcd(cfg)
	assert.NoError(t, err)

	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{clURL},
	})

	assert.NoError(t, err)

	return client, func() {
		if e != nil {
			assert.NoError(t, e.Shutdown())
		}
		remove()
	}
}

func TestEnqueue(t *testing.T) {
	t.Parallel()
	client, cleanup := newEtcd(t)
	defer cleanup()
	queue := New("testenq", client)
	err := queue.Enqueue(context.Background(), "test value")
	assert.NoError(t, err)
}

func TestDequeue(t *testing.T) {
	t.Parallel()
	client, cleanup := newEtcd(t)
	defer cleanup()

	queue := New("testdeq", client)
	err := queue.Enqueue(context.Background(), "test value")
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
	client, cleanup := newEtcd(t)
	defer cleanup()

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
	client, cleanup := newEtcd(t)
	defer cleanup()

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
