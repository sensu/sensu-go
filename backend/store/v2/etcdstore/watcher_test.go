package etcdstore_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/etcd"
	oldstore "github.com/sensu/sensu-go/backend/store/etcd"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/tests/v3/integration"
)

const timeout = 10

func TestWatch(t *testing.T) {
	type storeFunc func(context.Context, storev2.Interface)
	foo := corev3.FixtureEntityConfig("foo")
	wrapper, err := storev2.WrapResource(foo)
	if err != nil {
		t.Fatal(err)
	}
	bar := corev3.FixtureEntityConfig("foo")
	bar.Metadata.Labels["foo"] = "bar"
	wrapper2, err := storev2.WrapResource(bar)
	if err != nil {
		t.Fatal(err)
	}
	req := storev2.NewResourceRequestFromResource(context.Background(), foo)
	req2 := storev2.NewResourceRequestFromResource(context.Background(), bar)

	testWithEtcdClient(t, func(s storev2.Interface, client *clientv3.Client) {
		tests := []struct {
			name         string
			storeFunc    storeFunc
			wantAction   storev2.WatchActionType
			wantResource corev3.Resource
		}{
			{
				name: "resource is created",
				storeFunc: func(ctx context.Context, s storev2.Interface) {
					if err := s.CreateOrUpdate(req, wrapper); err != nil {
						t.Fatal(err)
					}
				},
				wantAction:   storev2.WatchCreate,
				wantResource: foo,
			},
			{
				name: "resource is updated",
				storeFunc: func(ctx context.Context, s storev2.Interface) {
					if err := s.CreateOrUpdate(req2, wrapper2); err != nil {
						t.Fatal(err)
					}
				},
				wantAction:   storev2.WatchUpdate,
				wantResource: bar,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				req.Context = ctx

				w := s.Watch(req)

				var wg sync.WaitGroup
				wg.Add(1)
				go testCheckResult(t, w, tt.wantAction, tt.wantResource, &wg)
				time.Sleep(time.Second) // forgive me
				tt.storeFunc(ctx, s)
				wg.Wait()

				cancel()
				testCheckStoppedWatcher(t, w)
			})
		}
	})
}

func TestWatchRetry(t *testing.T) {
	integration.BeforeTestExternal(t)
	c := integration.NewClusterV3(t, &integration.ClusterConfig{GRPCKeepAliveInterval: 1 * time.Second, GRPCKeepAliveTimeout: 2 * time.Second, Size: 3})
	defer c.Terminate(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := c.Client(0)
	s := etcdstore.NewStore(client)
	oldStore := oldstore.NewStore(client)

	ns := &corev2.Namespace{Name: "default"}

	if err := oldStore.CreateNamespace(context.Background(), ns); err != nil {
		t.Fatal(err)
	}

	// Create resource
	foo := corev3.FixtureEntityConfig("foo")
	req := storev2.NewResourceRequestFromResource(ctx, foo)
	wrapper, err := storev2.WrapResource(foo)
	if err != nil {
		t.Fatal(err)
	}
	w := s.Watch(req)

	var wg sync.WaitGroup
	wg.Add(1)
	go testCheckResult(t, w, storev2.WatchCreate, foo, &wg)
	time.Sleep(time.Second)
	if err := s.CreateOrUpdate(req, wrapper); err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Create a failure with the watcher by partioning our etcd cluster
	c.Members[0].InjectPartition(t, c.Members[1], c.Members[2])
	time.Sleep(1 * time.Second)

	c.Members[0].RecoverPartition(t, c.Members[1], c.Members[2])
	c.Members[0].WaitOK(t)
	time.Sleep(1 * time.Second)

	// Update resource
	foo.Metadata.Labels["foo"] = "acme"
	wrapper, err = storev2.WrapResource(foo)
	if err != nil {
		t.Fatal(err)
	}
	var wg2 sync.WaitGroup
	wg2.Add(1)
	go testCheckResult(t, w, storev2.WatchUpdate, foo, &wg2)
	time.Sleep(time.Second)
	if err := s.CreateOrUpdate(req, wrapper); err != nil {
		t.Fatal(err)
	}
	wg2.Wait()
}

func testCheckResult(t *testing.T, w <-chan []storev2.WatchEvent, action storev2.WatchActionType, resource corev3.Resource, wg *sync.WaitGroup) {
	t.Helper()
	defer wg.Done()

	select {
	case events, ok := <-w:
		if !ok {
			panic("channel closed")
		}
		event := events[0]
		if event.Type != action {
			t.Errorf("event type = %v, want %v", event.Type, action)
		}

		if event.Type == storev2.WatchError {
			t.Error("watch error")
			return
		}

		var got corev3.EntityConfig
		if event.Value != nil {
			if err := event.Value.UnwrapInto(&got); err != nil {
				t.Error(err)
				return
			}
		} else if event.PreviousValue != nil {
			if err := event.Value.UnwrapInto(&got); err != nil {
				t.Error(err)
				return
			}
		}

		if !cmp.Equal(resource.GetMetadata(), got.GetMetadata()) {
			t.Errorf("watch result = %#v, want %#v", got, resource)
			return
		}
	case <-time.After(timeout * time.Second):
		t.Fatalf("timeout after waiting %d for the Result() chan", timeout)
	}
}

func testCheckStoppedWatcher(t *testing.T, w <-chan []storev2.WatchEvent) {
	t.Helper()

	select {
	case _, ok := <-w:
		if ok {
			t.Fatal("resultChan should have been closed")
		}
	case <-time.After(timeout * time.Second):
		t.Fatalf("timeout after waiting %d for resultChan", timeout)
	}
}

func testWithEtcdClient(t *testing.T, f func(storev2.Interface, *clientv3.Client)) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()

	s := etcdstore.NewStore(client)
	oldStore := oldstore.NewStore(client)
	ns := &corev2.Namespace{Name: "default"}

	if err := oldStore.CreateNamespace(context.Background(), ns); err != nil {
		t.Fatal(err)
	}

	f(s, client)
}
