// +build integration,!race

// These tests are unfortunately quite slow. This is somewhat mitigated by the
// fact that they are parallelized, but still consume 30-45 seconds. This is
// due to the fact that etcd lease expirations cannot be hurried. No, I don't
// want to mock them out. :)
package ringv2

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/etcd"
)

func TestAddWithoutSetInterval(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ring := New(client, t.Name())

	if err := ring.Add(context.Background(), "a"); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.SetInterval(ctx, 5); err != nil {
		t.Fatal(err)
	}

	if err := ring.Add(ctx, "foo"); err != nil {
		t.Fatal(err)
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.SetInterval(ctx, 5); err != nil {
		t.Fatal(err)
	}

	if err := ring.Add(context.TODO(), "foo"); err != nil {
		t.Fatal(err)
	}

	if err := ring.Remove(context.TODO(), "foo"); err != nil {
		t.Fatal(err)
	}
}

func TestWatchAddRemove(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.SetInterval(ctx, 5); err != nil {
		t.Fatal(err)
	}

	wc := ring.Watch(ctx)

	if err := ring.Add(ctx, "foo"); err != nil {
		t.Fatal(err)
	}

	got := <-wc

	want := Event{
		Type:  EventAdd,
		Value: "foo",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad event: got %v, want %v", got, want)
	}

	if err := ring.Remove(context.TODO(), "foo"); err != nil {
		t.Fatal(err)
	}

	got = <-wc

	want = Event{
		Type:  EventRemove,
		Value: "foo",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad event: got %v, want %v", got, want)
	}

	if empty, err := ring.IsEmpty(ctx); err != nil {
		t.Fatal(err)
	} else if !empty {
		t.Fatal("ring not empty but should be")
	}
}

func TestWatchTrigger(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ring := New(client, t.Name())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.SetInterval(ctx, 5); err != nil {
		t.Fatal(err)
	}

	wc := ring.Watch(ctx)

	if err := ring.Add(ctx, "foo"); err != nil {
		t.Fatal(err)
	}

	// Drain the add event
	<-wc

	for i := 0; i < 2; i++ {
		got := <-wc
		want := Event{
			Type:  EventTrigger,
			Value: "foo",
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("bad event: got %v, want %v", got, want)
		}
	}
}

func TestRingOrdering(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ring := New(client, t.Name())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.SetInterval(ctx, 5); err != nil {
		t.Fatal(err)
	}

	wc := ring.Watch(ctx)

	items := []string{
		"mulder", "scully", "skinner",
	}

	for _, item := range items {
		if err := ring.Add(ctx, item); err != nil {
			t.Fatal(err)
		}
	}

	if empty, err := ring.IsEmpty(ctx); err != nil {
		t.Fatal(err)
	} else if empty {
		t.Fatal("ring empty but shouldn't be")
	}

	for range items {
		// Drain the add events
		<-wc
	}

	for i := 0; i < 5; i++ {
		got := <-wc
		want := Event{
			Type:  EventTrigger,
			Value: items[i%len(items)],
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("bad event: got %v, want %v", got, want)
		}
	}
}

func TestConcurrentRingOrdering(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.SetInterval(ctx, 5); err != nil {
		t.Fatal(err)
	}

	wc1 := ring.Watch(ctx)
	wc2 := ring.Watch(ctx)
	wc3 := ring.Watch(ctx)

	items := []string{
		"mulder", "scully", "skinner",
	}

	for _, item := range items {
		if err := ring.Add(ctx, item); err != nil {
			t.Fatal(err)
		}
	}

	if empty, err := ring.IsEmpty(ctx); err != nil {
		t.Fatal(err)
	} else if empty {
		t.Fatal("ring empty but shouldn't be")
	}

	for i := range items {
		// Drain the add events
		for _, wc := range []<-chan Event{wc1, wc2, wc3} {
			got := <-wc

			want := Event{
				Type:  EventAdd,
				Value: items[i],
			}

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("bad event: got %v, want %v", got, want)
			}
		}
	}

	events := make([][]Event, 3)
	var wg sync.WaitGroup
	wg.Add(3)

	for i, wc := range []<-chan Event{wc1, wc2, wc3} {
		go func(wc <-chan Event, i int) {
			for j := 0; j < 5; j++ {
				events[i] = append(events[i], <-wc)
			}
			wg.Done()
		}(wc, i)
	}

	wg.Wait()

	exp := []Event{
		{Type: EventTrigger, Value: "mulder"},
		{Type: EventTrigger, Value: "scully"},
		{Type: EventTrigger, Value: "skinner"},
		{Type: EventTrigger, Value: "mulder"},
		{Type: EventTrigger, Value: "scully"},
	}

	for i := range events {
		t.Run(fmt.Sprintf("client %d", i), func(t *testing.T) {
			if got, want := events[i], exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad events: got %v, want %v", got, want)
			}
		})
	}
}

// eventTest tests that for a given sequence of Add and Remove actions,
// a certain sequence of events is observed. eventTest will use the presence
// of EventAdd and EventRemove events to call ring.Add and ring.Remove.
//
// Given a sequence of want Events, eventTest will execute EventAdd and
// EventRemove, and attempt to observe EventTrigger.
func eventTest(t *testing.T, want []Event) {
	t.Helper()
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.SetInterval(ctx, 5); err != nil {
		t.Fatal(err)
	}

	wc := ring.Watch(ctx)

	var got []Event

	for _, event := range want {
		switch event.Type {
		case EventAdd:
			if err := ring.Add(ctx, event.Value); err != nil {
				t.Fatal(err)
			}
		case EventRemove:
			if err := ring.Remove(ctx, event.Value); err != nil {
				t.Fatal(err)
			}
		}
		got = append(got, <-wc)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad events: got %v, want %v", got, want)
	}
}

func TestRemoveNextTrigger(t *testing.T) {
	eventTest(t, []Event{
		{Type: EventAdd, Value: "mulder"},
		{Type: EventAdd, Value: "scully"},
		{Type: EventAdd, Value: "skinner"},
		{Type: EventTrigger, Value: "mulder"},
		{Type: EventTrigger, Value: "scully"},
		{Type: EventRemove, Value: "skinner"},
		{Type: EventTrigger, Value: "mulder"},
	})
}

func TestWatchAndAddAfter(t *testing.T) {
	eventTest(t, []Event{
		{Type: EventAdd, Value: "byers"},
		{Type: EventAdd, Value: "frohike"},
		{Type: EventTrigger, Value: "byers"},
		{Type: EventAdd, Value: "langly"},
		{Type: EventTrigger, Value: "frohike"},
		{Type: EventTrigger, Value: "byers"},
		{Type: EventTrigger, Value: "langly"},
	})
}

func TestWatchAfterAdd(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.SetInterval(ctx, 5); err != nil {
		t.Fatal(err)
	}

	if err := ring.Add(ctx, "fowley"); err != nil {
		t.Fatal(err)
	}

	wc := ring.Watch(ctx)

	got := <-wc
	want := Event{Type: EventTrigger, Value: "fowley"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad event: got %v, want %v", got, want)
	}
}

func GetSetInterval(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.SetInterval(ctx, 5); err != nil {
		t.Fatal(err)
	}

	wc := ring.Watch(ctx)

	if err := ring.Add(ctx, "covarrubias"); err != nil {
		t.Fatal(err)
	}

	// drain add event
	<-wc

	start := time.Now()

	if err := ring.SetInterval(ctx, 10); err != nil {
		t.Fatal(err)
	}

	// drain trigger event
	<-wc

	// >10s should have elapsed
	if time.Now().Sub(start) < (10 * time.Second) {
		t.Fatal("ineffectual SetInterval")
	}
}

func TestLeaseExpiryWithNoWatcher(t *testing.T) {
	// This test reaches into the implementation details of the ring in order
	// to observe trigger events without Watch(). If the implementation of
	// the ring changes, this test will be invalidated and should be removed
	// or changed.
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.SetInterval(ctx, 5); err != nil {
		t.Fatal(err)
	}

	wc := ring.Watch(ctx)

	customCtx, customCancel := context.WithCancel(context.Background())
	defer customCancel()

	triggerWatch := ring.client.Watch(customCtx, ring.triggerPrefix, clientv3.WithPrefix(), clientv3.WithFilterPut())

	if err := ring.Add(ctx, "cgb"); err != nil {
		t.Fatal(err)
	}

	// drain add event
	<-wc

	// cancel the watcher, so that nothing is handling the lease expiration
	cancel()

	<-triggerWatch

	wc = ring.Watch(customCtx)

	got := <-wc
	want := Event{Type: EventTrigger, Value: "cgb"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad event: got %v, want %v", got, want)
	}
}
