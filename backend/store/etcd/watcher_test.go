package etcd

import (
	"context"
	"reflect"
	"sync"
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
			ctx:               context.Background(),
			client:            client,
			key:               EtcdRoot,
			recursive:         true,
			incomingEventChan: make(chan store.WatchEvent, incomingEventChanBufSize),
			resultChan:        make(chan store.WatchEvent, resultChanBufSize),
			logger:            logger,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		opts := []clientv3.OpOption{clientv3.WithCreatedNotify(), clientv3.WithPrefix()}
		watchChanStopped := make(chan struct{})
		w.watch(ctx, opts, watchChanStopped)

		if err := client.ActiveConnection().Close(); err != nil {
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
			ctx:               ctx,
			cancel:            cancel,
			client:            client,
			key:               EtcdRoot,
			recursive:         true,
			incomingEventChan: make(chan store.WatchEvent, incomingEventChanBufSize),
			resultChan:        make(chan store.WatchEvent, resultChanBufSize),
			logger:            logger,
		}

		w.start()
		cancel()

		testCheckStoppedWatcher(t, w)
	})
}

func TestWatchProcessEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	nullLogger, hook := test.NewNullLogger()
	w := &Watcher{
		ctx:               ctx,
		cancel:            cancel,
		incomingEventChan: make(chan store.WatchEvent, 2),
		resultChan:        make(chan store.WatchEvent, 1),
		logger:            nullLogger.WithField("testing", true),
	}

	var wg sync.WaitGroup
	go w.processEvents(&wg)

	// Inject a first event, we should not have any log entry
	w.incomingEventChan <- store.WatchEvent{}
	if len(hook.Entries) != 0 {
		t.Errorf("expected no log entries, got %d", len(hook.Entries))
	}

	// Inject a second event, which should block because the resultChan buffer is
	// full. We should therefore receive a warning
	go func() {
		w.incomingEventChan <- store.WatchEvent{}
	}()
	time.Sleep(1 * time.Second)
	if len(hook.Entries) != 1 {
		t.Errorf("expected one log entry, got %d", len(hook.Entries))
	}

	cancel()
	wg.Wait()
}

func TestWatchQueueEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	nullLogger, hook := test.NewNullLogger()
	w := &Watcher{
		ctx:               ctx,
		incomingEventChan: make(chan store.WatchEvent, 1),
		logger:            nullLogger.WithField("testing", true),
	}

	// Queue a first event, we should not have any log entry
	w.queueEvent(ctx, store.WatchEvent{})

	// Inject a second event, which should block because the resultChan buffer is
	// full. We should therefore receive a warning
	go func() {
		w.queueEvent(ctx, store.WatchEvent{})
	}()
	time.Sleep(1 * time.Second)
	if len(hook.Entries) != 1 {
		t.Errorf("expected one log entry, got %d", len(hook.Entries))
	}

	cancel()
}

func TestWatchRetry(t *testing.T) {
	c := integration.NewClusterV3(t, &integration.ClusterConfig{GRPCKeepAliveInterval: 1 * time.Second, GRPCKeepAliveTimeout: 2 * time.Second, Size: 3})
	defer c.Terminate(t)

	ctx, cancel := context.WithCancel(context.Background())

	client0 := c.Client(0)
	store0 := NewStore(client0, "store0")
	w := Watch(ctx, client0, "/sensu.io", true)

	// Create resource
	foo := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "foo"}}
	if err := store0.CreateOrUpdateResource(ctx, foo); err != nil {
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
	if err := store0.CreateOrUpdateResource(ctx, foo); err != nil {
		t.Fatal(err)
	}

	testCheckResult(t, w, store.WatchUpdate, foo)
	cancel()
}

func testCheckResult(t *testing.T, w *Watcher, action store.WatchActionType, resource corev2.Resource) {
	t.Helper()

	select {
	case event := <-w.Result():
		if event.Type != action {
			t.Errorf("event type = %v, want %v", event.Type, action)
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
