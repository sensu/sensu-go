// +build integration,!race

package etcd

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/integration"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/fixture"
	"github.com/sirupsen/logrus/hooks/test"
)

const timeout = 10

func TestWatch(t *testing.T) {
	type storeFunc func(context.Context, store.Store)
	foo := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "foo"}}
	fooBis := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "foo"}, Foo: "acme"}

	testWithEtcdClient(t, func(s store.Store, client *clientv3.Client) {
		tests := []struct {
			name         string
			key          string
			storeFunc    storeFunc
			wantAction   store.WatchActionType
			wantResource corev2.Resource
		}{
			{
				name: "resource is created",
				key:  EtcdRoot + "/" + foo.StorePrefix(),
				storeFunc: func(ctx context.Context, s store.Store) {
					if err := s.CreateOrUpdateResource(ctx, foo); err != nil {
						t.Fatal(err)
					}
				},
				wantAction:   store.WatchCreate,
				wantResource: foo,
			},
			{
				name: "resource is updated",
				key:  EtcdRoot + "/" + fooBis.StorePrefix(),
				storeFunc: func(ctx context.Context, s store.Store) {
					if err := s.CreateOrUpdateResource(ctx, fooBis); err != nil {
						t.Fatal(err)
					}
				},
				wantAction:   store.WatchUpdate,
				wantResource: fooBis,
			},
			{
				name: "resource is deleted",
				key:  EtcdRoot + "/" + foo.StorePrefix(),
				storeFunc: func(ctx context.Context, s store.Store) {
					if err := s.DeleteResource(ctx, foo.StorePrefix(), foo.GetObjectMeta().Name); err != nil {
						t.Fatal(err)
					}
				},
				wantAction:   store.WatchDelete,
				wantResource: fooBis,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				w := Watch(ctx, client, tt.key, true)
				tt.storeFunc(ctx, s)

				testCheckResult(t, w, tt.wantAction, tt.wantResource)

				cancel()
				testCheckStoppedWatcher(t, w)
			})
		}
	})
}

func TestWatchErrConnClosed(t *testing.T) {
	testWithEtcdClient(t, func(s store.Store, client *clientv3.Client) {
		w := &Watcher{
			ctx:        context.Background(),
			client:     client,
			key:        EtcdRoot,
			recursive:  true,
			resultChan: make(chan store.WatchEvent, resultChanBufSize),
			logger:     logger,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		opts := []clientv3.OpOption{clientv3.WithCreatedNotify(), clientv3.WithPrefix()}
		watchChanStopped := make(chan struct{})
		w.watch(ctx, opts, watchChanStopped)

		if err := client.Close(); err != nil && err != context.Canceled {
			t.Fatal(err)
		}

		select {
		case <-time.After(timeout * time.Second):
			t.Fatalf("timeout after waiting %d for resultChan", timeout)
		case <-watchChanStopped:
		}
	})
}

func TestWatchContextCancel(t *testing.T) {
	testWithEtcdClient(t, func(s store.Store, client *clientv3.Client) {
		ctx, cancel := context.WithCancel(context.Background())

		w := &Watcher{
			ctx:        ctx,
			cancel:     cancel,
			client:     client,
			key:        EtcdRoot,
			recursive:  true,
			resultChan: make(chan store.WatchEvent, resultChanBufSize),
			logger:     logger,
		}

		w.start()
		cancel()

		testCheckStoppedWatcher(t, w)
	})
}

func TestWatchQueueEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	nullLogger, hook := test.NewNullLogger()
	w := &Watcher{
		ctx:        ctx,
		resultChan: make(chan store.WatchEvent, 1),
		logger:     nullLogger.WithField("testing", true),
	}

	// Queue a first event, we should not have any log entry
	w.queueEvent(ctx, store.WatchEvent{})

	// Inject a second event, which should block because the resultChan buffer is
	// full. We should therefore receive a warning
	go func() {
		w.queueEvent(ctx, store.WatchEvent{})
	}()
	time.Sleep(1 * time.Second)
	if len(hook.AllEntries()) != 1 {
		t.Errorf("expected one log entry, got %d", len(hook.Entries))
	}

	cancel()
}

func TestWatchRetry(t *testing.T) {
	c := integration.NewClusterV3(t, &integration.ClusterConfig{GRPCKeepAliveInterval: 1 * time.Second, GRPCKeepAliveTimeout: 2 * time.Second, Size: 3})
	defer c.Terminate(t)

	ctx, cancel := context.WithCancel(context.Background())

	client := c.Client(0)
	s := NewStore(client, "store0")
	w := Watch(ctx, client, "/sensu.io", true)

	// Create resource
	foo := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "foo"}}
	if err := s.CreateOrUpdateResource(ctx, foo); err != nil {
		t.Fatal(err)
	}
	testCheckResult(t, w, store.WatchCreate, foo)

	// Create a failure with the watcher by partioning our etcd cluster
	c.Members[0].InjectPartition(t, c.Members[1], c.Members[2])
	time.Sleep(1 * time.Second)

	c.Members[0].RecoverPartition(t, c.Members[1], c.Members[2])
	c.Members[0].WaitOK(t)
	time.Sleep(1 * time.Second)

	// Update resource
	foo.Foo = "acme"
	if err := s.CreateOrUpdateResource(ctx, foo); err != nil {
		t.Fatal(err)
	}

	testCheckResult(t, w, store.WatchUpdate, foo)
	cancel()
}

func TestWatchCompactedRevision(t *testing.T) {
	c := integration.NewClusterV3(t, &integration.ClusterConfig{GRPCKeepAliveInterval: 1 * time.Second, GRPCKeepAliveTimeout: 2 * time.Second, Size: 3})
	defer c.Terminate(t)

	ctx, cancel := context.WithCancel(context.Background())

	client := c.Client(0)
	s := NewStore(client, "store")

	// Create the 'foo' resource to generate a new revision (revision=2)
	foo := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "foo"}}
	if err := s.CreateOrUpdateResource(ctx, foo); err != nil {
		t.Fatal(err)
	}

	// Update the 'foo' resource to generate an additional revision (revision=3)
	foo.Foo = "acme"
	if err := s.CreateOrUpdateResource(ctx, foo); err != nil {
		t.Fatal(err)
	}

	// Compact to the latest revision
	client.Compact(ctx, int64(3), clientv3.WithCompactPhysical())

	// Start a watcher with a compacted revision (1)
	w := &Watcher{
		ctx:        context.Background(),
		client:     client,
		key:        EtcdRoot,
		recursive:  true,
		revision:   int64(1),
		resultChan: make(chan store.WatchEvent, resultChanBufSize),
		logger:     logger,
	}
	w.start()

	// The watcher should return the event associated with the latest revision
	// (3), which corresponds to the update to the foo resource
	testCheckResult(t, w, store.WatchError, nil)
	testCheckResult(t, w, store.WatchUpdate, foo)

	// Since we handled the revision 3, if the watcher was to be restarted, it
	// should pick up from revision 4
	if w.revision != 4 {
		t.Errorf("watcher revision = %d, want %d", w.revision, 4)
	}

	cancel()
}

func TestWatchRevisions(t *testing.T) {
	c := integration.NewClusterV3(t, &integration.ClusterConfig{GRPCKeepAliveInterval: 1 * time.Second, GRPCKeepAliveTimeout: 2 * time.Second, Size: 3})
	defer c.Terminate(t)

	ctx, cancel := context.WithCancel(context.Background())

	client := c.Client(0)
	s := NewStore(client, "store")

	// Create the 'foo' resource to generate a new revision (revision=2)
	foo := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "foo"}}
	if err := s.CreateOrUpdateResource(ctx, foo); err != nil {
		t.Fatal(err)
	}

	// Update the 'foo' resource to generate an additional revision (revision=3)
	fooBis := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "foo"}}
	fooBis.Foo = "acme"
	if err := s.CreateOrUpdateResource(ctx, fooBis); err != nil {
		t.Fatal(err)
	}

	// Start a watcher with a compacted revision (1)
	w := &Watcher{
		ctx:        context.Background(),
		client:     client,
		key:        EtcdRoot,
		recursive:  true,
		revision:   int64(1),
		resultChan: make(chan store.WatchEvent, resultChanBufSize),
		logger:     logger,
	}
	w.start()

	// The watcher should return all events from revision 1
	testCheckResult(t, w, store.WatchCreate, foo)
	testCheckResult(t, w, store.WatchUpdate, fooBis)

	// Since we handled the revision 3, if the watcher was to be restarted, it
	// should pick up from revision 4
	if w.revision != 4 {
		t.Errorf("watcher revision = %d, want %d", w.revision, 4)
	}

	// Update the 'foo' resource to generate an additional revision (revision=4)
	fooBis.Foo = "bar"
	if err := s.CreateOrUpdateResource(ctx, fooBis); err != nil {
		t.Fatal(err)
	}

	// The watcher should also return this new watch event
	testCheckResult(t, w, store.WatchUpdate, fooBis)

	// The tracked revision should also be bumped
	if w.revision != 5 {
		t.Errorf("watcher revision = %d, want %d", w.revision, 5)
	}

	cancel()
}

func testCheckResult(t *testing.T, w *Watcher, action store.WatchActionType, resource corev2.Resource) {
	t.Helper()

	select {
	case event := <-w.Result():
		if event.Type != action {
			t.Errorf("event type = %v, want %v", event.Type, action)
		}

		if event.Type == store.WatchError {
			return
		}

		got := &fixture.Resource{}
		if err := unmarshal(event.Object, got); err != nil {
			t.Errorf("could not decode event object: %v", err)
			return
		}

		if !reflect.DeepEqual(resource, got) {
			t.Errorf("watch result = %#v, want %#v", got, resource)
			return
		}
	case <-time.After(timeout * time.Second):
		t.Fatalf("timeout after waiting %d for the Result() chan", timeout)
	}
}

func testCheckStoppedWatcher(t *testing.T, w *Watcher) {
	t.Helper()

	select {
	case _, ok := <-w.resultChan:
		if ok {
			t.Fatal("resultChan should have been closed")
		}
	case <-time.After(timeout * time.Second):
		t.Fatalf("timeout after waiting %d for resultChan", timeout)
	}
}
