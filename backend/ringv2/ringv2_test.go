//go:build integration && !race
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

	"github.com/sensu/sensu-go/backend/etcd"
)

func TestAdd(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.Add(ctx, "foo", 600); err != nil {
		t.Fatal(err)
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.Add(ctx, "foo", 600); err != nil {
		t.Fatal(err)
	}

	if err := ring.Remove(ctx, "foo"); err != nil {
		t.Fatal(err)
	}
}

func TestWatchAddRemove(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wc := ring.Watch(ctx, "test", 1, 5, "")

	if err := ring.Add(ctx, "foo", 600); err != nil {
		t.Fatal(err)
	}

	got := <-wc

	want := Event{
		Type:   EventAdd,
		Values: []string{"foo"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad event: got %v, want %v", got, want)
	}

	if err := ring.Remove(context.TODO(), "foo"); err != nil {
		t.Fatal(err)
	}

	got = <-wc

	want = Event{
		Type:   EventRemove,
		Values: []string{"foo"},
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

	client := e.NewEmbeddedClient()
	defer client.Close()

	ring := New(client, t.Name())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wc := ring.Watch(ctx, "test", 1, 5, "")

	if err := ring.Add(ctx, "foo", 600); err != nil {
		t.Fatal(err)
	}

	// Drain the add event
	<-wc

	for i := 0; i < 2; i++ {
		got := <-wc
		want := Event{
			Type:   EventTrigger,
			Values: []string{"foo"},
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

	client := e.NewEmbeddedClient()
	defer client.Close()

	ring := New(client, t.Name())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wc := ring.Watch(ctx, "test", 1, 5, "")

	items := []string{
		"mulder", "scully", "skinner",
	}

	for _, item := range items {
		if err := ring.Add(ctx, item, 600); err != nil {
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
			Type:   EventTrigger,
			Values: []string{items[i%len(items)]},
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

	client := e.NewEmbeddedClient()
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wc1 := ring.Watch(ctx, "test1", 1, 5, "")
	wc2 := ring.Watch(ctx, "test2", 1, 5, "")
	wc3 := ring.Watch(ctx, "test3", 1, 5, "")

	items := []string{
		"mulder", "scully", "skinner",
	}

	for _, item := range items {
		if err := ring.Add(ctx, item, 600); err != nil {
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
				Type:   EventAdd,
				Values: []string{items[i]},
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
		{Type: EventTrigger, Values: []string{"mulder"}},
		{Type: EventTrigger, Values: []string{"scully"}},
		{Type: EventTrigger, Values: []string{"skinner"}},
		{Type: EventTrigger, Values: []string{"mulder"}},
		{Type: EventTrigger, Values: []string{"scully"}},
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

	client := e.NewEmbeddedClient()
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wc := ring.Watch(ctx, "test", 1, 5, "")

	var got []Event

	for _, event := range want {
		switch event.Type {
		case EventAdd:
			if err := ring.Add(ctx, event.Values[0], 600); err != nil {
				t.Fatal(err)
			}
		case EventRemove:
			if err := ring.Remove(ctx, event.Values[0]); err != nil {
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
		{Type: EventAdd, Values: []string{"mulder"}},
		{Type: EventAdd, Values: []string{"scully"}},
		{Type: EventAdd, Values: []string{"skinner"}},
		{Type: EventTrigger, Values: []string{"mulder"}},
		{Type: EventTrigger, Values: []string{"scully"}},
		{Type: EventRemove, Values: []string{"skinner"}},
		{Type: EventTrigger, Values: []string{"mulder"}},
	})
}

func TestWatchAndAddAfter(t *testing.T) {
	eventTest(t, []Event{
		{Type: EventAdd, Values: []string{"byers"}},
		{Type: EventAdd, Values: []string{"frohike"}},
		{Type: EventTrigger, Values: []string{"byers"}},
		{Type: EventAdd, Values: []string{"langly"}},
		{Type: EventTrigger, Values: []string{"frohike"}},
		{Type: EventTrigger, Values: []string{"langly"}},
		{Type: EventTrigger, Values: []string{"byers"}},
	})
}

func TestWatchAfterAdd(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ring.Add(ctx, "fowley", 600); err != nil {
		t.Fatal(err)
	}

	wc := ring.Watch(ctx, "test", 1, 5, "")

	got := <-wc
	want := Event{Type: EventTrigger, Values: []string{"fowley"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad event: got %v, want %v", got, want)
	}
}

func TestGetSetInterval(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wc := ring.Watch(ctx, "test", 1, 5, "")

	if err := ring.Add(ctx, "covarrubias", 600); err != nil {
		t.Fatal(err)
	}

	// drain add event
	<-wc

	start := time.Now()

	// drain trigger event
	<-wc

	// >10s should have elapsed
	if time.Since(start) < (5 * time.Second) {
		t.Fatal("ineffectual SetInterval")
	}
}

func TestMultipleItems(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	ring := New(client, t.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wc := ring.Watch(ctx, "test", 3, 5, "")

	items := []string{"byers", "frohike", "mulder", "scully", "skinner"}

	for _, item := range items {
		if err := ring.Add(ctx, item, 600); err != nil {
			t.Fatal(err)
		}
		// drain add event
		<-wc
	}

	event := <-wc

	if got, want := event.Values, []string{"byers", "frohike", "mulder"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad values: got %v, want %v", got, want)
	}

	event = <-wc

	if got, want := event.Values, []string{"scully", "skinner", "byers"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad values: got %v, want %v", got, want)
	}

	event = <-wc

	if got, want := event.Values, []string{"frohike", "mulder", "scully"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad values: got %v, want %v", got, want)
	}
}

// TestReWatch asserts that cancelling the context associated with one
// watcher does not affect other identical subscriptions
func TestReWatch(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	ring := New(client, t.Name())
	testCtx, testCancel := context.WithCancel(context.Background())
	defer testCancel()

	ctx, cancel := context.WithCancel(testCtx)
	wc := ring.Watch(ctx, "test", 1, 5, "")

	if err := ring.Add(testCtx, "foo", 100); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		e := <-wc
		t.Logf("got event %v", e)
		c := 0
		for e.Type != EventTrigger {
			if c > 5 {
				t.Fatal("caught in error loop")
			}
			e = <-wc
			c++
		}
		cancel()
		ctx, cancel = context.WithCancel(testCtx)
		wc = ring.Watch(ctx, "test", 1, 5, "")
	}

	cancel()
}
