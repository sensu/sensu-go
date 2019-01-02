package liveness

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var logger = logrus.New()

func TestSwitchSet(t *testing.T) {
	// This test sets up a SwitchSet and two callbacks that are used
	// to verify that the SwitchSet is working as expected. It is
	// expected to function deterministically, and is not dependent on timing
	// or filesystem latency.
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var aliveWG sync.WaitGroup
	var deadWG sync.WaitGroup
	results := make(map[string][]int)

	// This callback gets executed when the entity dies
	expired := func(key string, prev State, revision int64) {
		mu.Lock()
		defer mu.Unlock()
		results[key] = append(results[key], 1)
		deadWG.Done()
	}

	// This callback gets executed when the entity asserts its liveness
	alive := func(key string, prev State, revision int64) {
		mu.Lock()
		defer mu.Unlock()
		results[key] = append(results[key], 0)
		aliveWG.Done()
	}

	toggle := NewSwitchSet(client, "test", expired, alive, logger)
	go toggle.Monitor(ctx)

	// the [0, 0, 0, 1] sequences
	deadWG.Add(1)
	aliveWG.Add(3)
	for i := 0; i < 3; i++ {
		if err := toggle.Alive(ctx, "entity1", 5); err != nil {
			t.Fatal(err)
		}
	}
	aliveWG.Wait()
	deadWG.Wait()

	// The subsequent [0, 1, 1, 1] sequence
	deadWG.Add(3)
	aliveWG.Add(1)
	if err := toggle.Alive(ctx, "entity1", 5); err != nil {
		t.Fatal(err)
	}
	aliveWG.Wait()
	deadWG.Wait()

	mu.Lock()
	if got, want := results["entity1"], []int{0, 0, 0, 1, 0, 1, 1, 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("bad results: got %v, want %v", got, want)
	}
	mu.Unlock()
}

func TestDead(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var aliveWG sync.WaitGroup
	var deadWG sync.WaitGroup
	results := make(map[string][]int)

	// This callback gets executed when the entity dies
	expired := func(key string, prev State, revision int64) {
		mu.Lock()
		defer mu.Unlock()
		results[key] = append(results[key], 1)
		deadWG.Done()
	}

	// This callback gets executed when the entity asserts its liveness
	alive := func(key string, prev State, revision int64) {
		mu.Lock()
		defer mu.Unlock()
		results[key] = append(results[key], 0)
		aliveWG.Done()
	}

	toggle := NewSwitchSet(client, "test", expired, alive, logger)
	go toggle.Monitor(ctx)

	deadWG.Add(3)
	if err := toggle.Dead(ctx, "entity1", 5); err != nil {
		t.Fatal(err)
	}

	deadWG.Wait()

	mu.Lock()
	if got, want := results["entity1"], []int{1, 1, 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("bad results: got %v, want %v", got, want)
	}
	mu.Unlock()
}
