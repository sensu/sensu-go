//go:build integration
// +build integration

package liveness

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func TestSwitchSet(t *testing.T) {
	// This test sets up a SwitchSet and two callbacks that are used
	// to verify that the SwitchSet is working as expected. It is
	// expected to function deterministically, and is not dependent on timing
	// or filesystem latency.
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	aliveC := make(chan struct{})
	deadC := make(chan struct{})
	results := make(map[string][]int)

	// This callback gets executed when the entity dies
	expired := func(key string, prev State, leader bool) bool {
		mu.Lock()
		defer mu.Unlock()
		results[key] = append(results[key], int(Dead))
		deadC <- struct{}{}
		return false
	}

	// This callback gets executed when the entity asserts its liveness
	alive := func(key string, prev State, leader bool) bool {
		mu.Lock()
		defer mu.Unlock()
		results[key] = append(results[key], int(Alive))
		aliveC <- struct{}{}
		return false
	}

	toggle := NewSwitchSet(client, "test", expired, alive, logger)
	toggle.monitor(ctx)

	// the [0, 0, 0, 1] sequences
	for i := 0; i < 3; i++ {
		if err := toggle.Alive(ctx, "entity1", 5); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 3; i++ {
		<-aliveC
	}
	<-deadC

	// The subsequent [0, 1, 1, 1] sequence
	if err := toggle.Alive(ctx, "entity1", 5); err != nil {
		t.Fatal(err)
	}
	<-aliveC
	for i := 0; i < 3; i++ {
		<-deadC
	}

	mu.Lock()
	if got, want := results["entity1"], []int{0, 0, 0, 1, 0, 1, 1, 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("bad results: got %v, want %v", got, want)
	}
	mu.Unlock()
}

func TestDead(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	results := make(map[string][]int)
	aliveC := make(chan struct{})
	deadC := make(chan struct{})

	// This callback gets executed when the entity dies
	expired := func(key string, prev State, leader bool) bool {
		mu.Lock()
		defer mu.Unlock()
		results[key] = append(results[key], int(Dead))
		deadC <- struct{}{}
		return false
	}

	// This callback gets executed when the entity asserts its liveness
	alive := func(key string, prev State, leader bool) bool {
		mu.Lock()
		defer mu.Unlock()
		results[key] = append(results[key], int(Alive))
		aliveC <- struct{}{}
		return false
	}

	toggle := NewSwitchSet(client, "test", expired, alive, logger)
	toggle.monitor(ctx)

	if err := toggle.Dead(ctx, "entity1", 5); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		<-deadC
	}

	mu.Lock()
	if got, want := results["entity1"], []int{1, 1, 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("bad results: got %v, want %v", got, want)
	}
	mu.Unlock()
}

func TestBury(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This callback gets executed when the entity dies
	expired := func(key string, prev State, leader bool) bool {
		t.Fatal("expired should never be called on key " + key)
		return false
	}

	// This callback gets executed when the entity asserts its liveness
	alive := func(key string, prev State, leader bool) bool {
		t.Fatal("alive should never be called on key " + key)
		return false
	}

	toggle := NewSwitchSet(client, "test", expired, alive, logger)
	toggle.monitor(ctx)

	if err := toggle.Dead(ctx, "default/entity1", 5); err != nil {
		t.Fatal(err)
	}

	if err := toggle.Bury(ctx, "default/entity1"); err != nil {
		t.Fatal(err)
	}

	// Ensure that key expiration doesn't occur
	time.Sleep(6 * time.Second)
}

func TestBuryOnCallback(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This callback gets executed when the entity dies
	expired := func(key string, prev State, leader bool) bool {
		t.Fatal("expired should not have been called")
		return true
	}

	// This callback gets executed when the entity asserts its liveness
	alive := func(key string, prev State, leader bool) bool {
		if prev != Dead {
			t.Fatal("bad previous state")
		}
		return true
	}

	toggle := NewSwitchSet(client, "test", expired, alive, logger)
	toggle.monitor(ctx)

	if err := toggle.Alive(ctx, "default/entity1", 5); err != nil {
		t.Fatal(err)
	}

	// Ensure that the expired callback never fires
	time.Sleep(6 * time.Second)
}
