// +build integration,!race

package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

const timeout = 10

func TestWatcher(t *testing.T) {
	testWithEtcdClient(t, func(s store.Store, client *clientv3.Client) {
		// TODO(palourde): Remove logrus output
		logrus.SetOutput(os.Stdout)
		logrus.SetLevel(logrus.DebugLevel)

		// Generate a fixture check
		check := v2.FixtureCheckConfig("foo")
		check.Annotations["channel"] = "#monitoring"
		check.Labels["region"] = "us-west-1"
		ctx := context.WithValue(context.Background(), types.NamespaceKey, check.Namespace)

		// Create this check in the store
		w := Watch(ctx, client, checkKeyBuilder.Build(""), true)
		if err := s.UpdateCheckConfig(ctx, check); err != nil {
			t.Fatalf("create failed: %v", err)
		}
		expectObject(t, w, store.WatchCreate, check)
		w.Stop()

		// Update this check
		w = Watch(ctx, client, checkKeyBuilder.Build(""), true)
		check.Interval = 30
		if err := s.UpdateCheckConfig(ctx, check); err != nil {
			t.Fatalf("updated failed: %v", err)
		}
		expectObject(t, w, store.WatchUpdate, check)
		w.Stop()

		// Generate a second fixture check
		check2 := v2.FixtureCheckConfig("bar")
		check2.Annotations["channel"] = "#monitoring"
		check2.Labels["region"] = "us-west-1"

		// Create this second check in the store
		w = Watch(ctx, client, checkKeyBuilder.Build(""), true)
		if err := s.UpdateCheckConfig(ctx, check2); err != nil {
			t.Fatalf("create failed: %v", err)
		}
		expectObject(t, w, store.WatchCreate, check2)
		w.Stop()

		// Compact the history of etcd and make sure we can still watch
		w = Watch(ctx, client, checkKeyBuilder.Build(""), true)
		_, err := client.Compact(ctx, 1, clientv3.WithCompactPhysical())
		if err != nil {
			t.Fatalf("error compacting: %v", err)
		}
		if err := s.UpdateCheckConfig(ctx, check); err != nil {
			t.Fatalf("updated failed: %v", err)
		}
		expectObject(t, w, store.WatchUpdate, check)
		w.Stop()

		// Delete a key
		w = Watch(ctx, client, checkKeyBuilder.Build(""), true)
		if err := s.DeleteCheckConfigByName(ctx, check.Name); err != nil {
			t.Fatalf("deletion failed: %v", err)
		}
		expectType(t, w, store.WatchDelete)
		w.Stop()

		// Test a canceled context
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		w = Watch(canceledCtx, client, checkKeyBuilder.Build(""), true)

		select {
		case _, ok := <-w.Result():
			if ok {
				t.Error("Result() chan should be closed")
			}
		case <-time.After(timeout * time.Second):
			t.Fatalf("timeout after waiting %d for the Result() chan", timeout)
		}
	})
}

func TestWatcherDoesNotBlock(t *testing.T) {
	testWithEtcdStore(t, func(s *Store) {
		ctx, cancel := context.WithCancel(context.Background())

		w := createWatcher(ctx, s.client, checkKeyBuilder.Build(""), true)

		// Send an error, which should block resultChan. Then, cancel the ctx, which
		// should freed up the blocking on resultChan and therefore cause the run()
		// goroutine to return
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			// startedWG is required by the run() goroutine but we can ignore it here
			var startedWG sync.WaitGroup
			startedWG.Add(1)
			w.run(&startedWG)
			wg.Done()
		}()
		w.errChan <- fmt.Errorf("error")
		cancel()
		wg.Wait()
	})
}

func TestEtcdClientClosed(t *testing.T) {
	testWithEtcdStore(t, func(s *Store) {
		w := createWatcher(context.Background(), s.client, checkKeyBuilder.Build(""), true)

		// Close the etcd client, which should send an error via the etcd watcher
		// chan, and then consume resultChan to ensure the error was reported, and
		// therefore cause the run() goroutine to return
		var wg sync.WaitGroup
		wg.Add(1)
		// startedWG is used to ensure the watcher was properly started before
		// trying to use any of its watcher
		var startedWG sync.WaitGroup
		startedWG.Add(1)
		go func() {
			w.run(&startedWG)
			wg.Done()
		}()
		startedWG.Wait()
		// Close the etcd client
		if err := w.client.Close(); err != nil {
			t.Fatal(err)
		}
		expectType(t, w, store.WatchError)
		wg.Wait()
	})
}

func expectType(t *testing.T, w store.Watcher, typ store.WatchActionType) {
	t.Helper()

	select {
	case event := <-w.Result():
		if event.Type != typ {
			t.Errorf("event type = %v, want %v", event.Type, typ)
		}
	case <-time.After(timeout * time.Second):
		t.Fatalf("timeout after waiting %d for the Result() chan", timeout)
	}
}

func expectObject(t *testing.T, w store.Watcher, typ store.WatchActionType, wanted *v2.CheckConfig) {
	t.Helper()

	select {
	case event := <-w.Result():
		if event.Type != typ {
			t.Errorf("event type = %v, want %v", event.Type, typ)
			return
		}

		if key := checkKeyBuilder.WithNamespace("default").Build(wanted.Name); event.Key != key {
			t.Errorf("event key = %v, want %v", event.Key, key)
			return
		}

		got := &v2.CheckConfig{}
		if err := json.Unmarshal(event.Object, got); err != nil {
			t.Errorf("could not decode event object: %v", err)
			return
		}

		if !reflect.DeepEqual(wanted, got) {
			t.Errorf("store.GetProvider = %v, want %v", got, wanted)
			return
		}
	case <-time.After(timeout * time.Second):
		t.Fatalf("timeout after waiting %d for the Result() chan", timeout)
	}
}
