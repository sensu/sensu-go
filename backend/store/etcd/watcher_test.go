// +build integration,!race

package etcd

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const timeout = 10

func TestWatcher(t *testing.T) {
	testWithEtcdClient(t, func(s store.Store, client *clientv3.Client) {
		// Generate a fixture check
		check := v2.FixtureCheckConfig("foo")
		check.Annotations["channel"] = "#monitoring"
		check.Labels["region"] = "us-west-1"
		ctx := context.WithValue(context.Background(), types.NamespaceKey, check.Namespace)
		cancelCtx, cancel := context.WithCancel(ctx)

		// Create this check in the store
		w := Watch(cancelCtx, client, checkKeyBuilder.Build(""), true)
		if err := s.UpdateCheckConfig(ctx, check); err != nil {
			t.Fatalf("create failed: %v", err)
		}
		expectObject(t, w, store.WatchCreate, check)
		cancel()

		cancelCtx, cancel = context.WithCancel(ctx)
		// Update this check
		w = Watch(cancelCtx, client, checkKeyBuilder.Build(""), true)
		check.Interval = 30
		if err := s.UpdateCheckConfig(ctx, check); err != nil {
			t.Fatalf("updated failed: %v", err)
		}
		expectObject(t, w, store.WatchUpdate, check)
		cancel()

		// Generate a second fixture check
		check2 := v2.FixtureCheckConfig("bar")
		check2.Annotations["channel"] = "#monitoring"
		check2.Labels["region"] = "us-west-1"

		cancelCtx, cancel = context.WithCancel(ctx)
		// Create this second check in the store
		w = Watch(cancelCtx, client, checkKeyBuilder.Build(""), true)
		if err := s.UpdateCheckConfig(ctx, check2); err != nil {
			t.Fatalf("create failed: %v", err)
		}
		expectObject(t, w, store.WatchCreate, check2)
		cancel()

		cancelCtx, cancel = context.WithCancel(ctx)
		// Compact the history of etcd and make sure we can still watch
		w = Watch(cancelCtx, client, checkKeyBuilder.Build(""), true)
		_, err := client.Compact(ctx, 1, clientv3.WithCompactPhysical())
		if err != nil {
			t.Fatalf("error compacting: %v", err)
		}
		if err := s.UpdateCheckConfig(ctx, check); err != nil {
			t.Fatalf("updated failed: %v", err)
		}
		expectObject(t, w, store.WatchUpdate, check)
		cancel()

		cancelCtx, cancel = context.WithCancel(ctx)
		// Delete a key
		w = Watch(cancelCtx, client, checkKeyBuilder.Build(""), true)
		if err := s.DeleteCheckConfigByName(ctx, check.Name); err != nil {
			t.Fatalf("deletion failed: %v", err)
		}
		expectType(t, w, store.WatchDelete)
		cancel()

		// Test context cancelation
		cancelCtx, cancel = context.WithCancel(ctx)
		cancel()
		w = Watch(cancelCtx, client, checkKeyBuilder.Build(""), true)

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

func TestEtcdClientClosed(t *testing.T) {
	testWithEtcdStore(t, func(s *Store) {
		w := Watch(context.Background(), s.client, checkKeyBuilder.Build(""), true)

		// Close the etcd client
		if err := w.client.Close(); err != nil {
			t.Fatal(err)
		}
		expectType(t, w, store.WatchError)
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
		if err := unmarshal(event.Object, got); err != nil {
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
