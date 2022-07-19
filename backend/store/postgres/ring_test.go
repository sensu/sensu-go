package postgres

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/ringv2"
)

func ringName(tname string) string {
	return fmt.Sprintf("/sensu.io/rings/default/%s", tname)
}

func TestAdd(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		listener := pq.NewListener(dsn, time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Cleanup(func() {
			_ = listener.UnlistenAll()
			_ = listener.Close()
		})
		bus := NewBus(ctx, listener)
		ring, err := NewRing(db, bus, ringName(t.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = ring.Close()
		})

		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		entityName := "foo"
		entity := corev2.FixtureEntity(entityName)
		namespace := corev3.FixtureNamespace(entity.Namespace)
		if err := namespaceStore.CreateOrUpdate(ctx, namespace); err != nil {
			t.Fatal(err)
		}
		if err := entityStore.UpdateEntity(ctx, entity); err != nil {
			t.Fatal(err)
		}
		if err := ring.Add(ctx, entityName, 600); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRemove(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		listener := pq.NewListener(dsn, time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Cleanup(func() {
			_ = listener.UnlistenAll()
			_ = listener.Close()
		})
		bus := NewBus(ctx, listener)
		ring, err := NewRing(db, bus, ringName(t.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = ring.Close()
		})

		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		entityName := "foo"
		entity := corev2.FixtureEntity(entityName)
		namespace := corev3.FixtureNamespace(entity.Namespace)
		if err := namespaceStore.CreateOrUpdate(ctx, namespace); err != nil {
			t.Fatal(err)
		}
		if err := entityStore.UpdateEntity(ctx, entity); err != nil {
			t.Fatal(err)
		}
		if err := ring.Add(ctx, entityName, 600); err != nil {
			t.Fatal(err)
		}

		if err := ring.Remove(ctx, entityName); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAddRemoveIsEmpty(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		listener := pq.NewListener(dsn, time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Cleanup(func() {
			_ = listener.UnlistenAll()
			_ = listener.Close()
		})
		bus := NewBus(ctx, listener)
		ring, err := NewRing(db, bus, ringName(t.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = ring.Close()
		})

		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		entityName := "foo"
		entity := corev2.FixtureEntity(entityName)
		namespace := corev3.FixtureNamespace(entity.Namespace)
		if err := namespaceStore.CreateOrUpdate(ctx, namespace); err != nil {
			t.Fatal(err)
		}
		if err := entityStore.UpdateEntity(ctx, entity); err != nil {
			t.Fatal(err)
		}
		if err := ring.Add(ctx, entityName, 600); err != nil {
			t.Fatal(err)
		}

		if err := ring.Remove(context.Background(), entityName); err != nil {
			t.Fatal(err)
		}

		if empty, err := ring.IsEmpty(ctx); err != nil {
			t.Fatal(err)
		} else if !empty {
			t.Fatal("ring not empty but should be")
		}
	})
}

func TestWatchTrigger(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		listener := pq.NewListener(dsn, time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Cleanup(func() {
			_ = listener.UnlistenAll()
			_ = listener.Close()
		})
		bus := NewBus(ctx, listener)
		ring, err := NewRing(db, bus, ringName(t.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = ring.Close()
		})

		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		sub := ringv2.Subscription{
			Name:             "test",
			Items:            1,
			IntervalSchedule: 5,
		}
		wc := ring.Subscribe(ctx, sub)

		entityName := "foo"
		entity := corev2.FixtureEntity(entityName)
		namespace := corev3.FixtureNamespace(entity.Namespace)
		if err := namespaceStore.CreateOrUpdate(ctx, namespace); err != nil {
			t.Fatal(err)
		}
		if err := entityStore.UpdateEntity(ctx, entity); err != nil {
			t.Fatal(err)
		}
		if err := ring.Add(ctx, entityName, 600); err != nil {
			t.Fatal(err)
		}

		// first event may or may not be empty
		<-wc

		got := <-wc
		want := ringv2.Event{
			Type:   ringv2.EventTrigger,
			Values: []string{entityName},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("bad event: got %v, want %v", got, want)
		}
	})
}

func TestRingOrdering(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		listener := pq.NewListener(dsn, time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Cleanup(func() {
			_ = listener.UnlistenAll()
			_ = listener.Close()
		})
		bus := NewBus(ctx, listener)
		ring, err := NewRing(db, bus, ringName(t.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = ring.Close()
		})

		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		sub := ringv2.Subscription{
			Name:             "test",
			Items:            1,
			IntervalSchedule: 5,
		}

		items := []string{
			"mulder", "scully", "skinner",
		}

		for _, item := range items {
			entity := corev2.FixtureEntity(item)
			namespace := corev3.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.CreateOrUpdate(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			if err := ring.Add(ctx, item, 600); err != nil {
				t.Fatal(err)
			}
		}

		if empty, err := ring.IsEmpty(ctx); err != nil {
			t.Fatal(err)
		} else if empty {
			t.Fatal("ring empty but shouldn't be")
		}

		wc := ring.Subscribe(ctx, sub)

		for i := 0; i < 5; i++ {
			got := <-wc
			want := ringv2.Event{
				Type:   ringv2.EventTrigger,
				Values: []string{items[i%len(items)]},
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("bad event (iteration %d): got %v, want %v", i, got, want)
			}
		}
	})
}

func TestConcurrentRingOrdering(t *testing.T) {
	t.Skip("Skipping")
	t.Parallel()

	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		listener := pq.NewListener(dsn, time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Cleanup(func() {
			_ = listener.UnlistenAll()
			_ = listener.Close()
		})
		bus := NewBus(ctx, listener)
		ring, err := NewRing(db, bus, ringName(t.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = ring.Close()
		})

		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		items := []string{
			"mulder", "scully", "skinner",
		}

		for _, item := range items {
			entity := corev2.FixtureEntity(item)
			namespace := corev3.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.CreateOrUpdate(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			if err := ring.Add(ctx, item, 600); err != nil {
				t.Fatal(err)
			}
		}

		if empty, err := ring.IsEmpty(ctx); err != nil {
			t.Fatal(err)
		} else if empty {
			t.Fatal("ring empty but shouldn't be")
		}

		sub1 := ringv2.Subscription{
			Name:             "test",
			Items:            1,
			IntervalSchedule: 5,
		}
		sub2 := ringv2.Subscription{
			Name:             "test",
			Items:            1,
			IntervalSchedule: 5,
		}
		sub3 := ringv2.Subscription{
			Name:             "test",
			Items:            1,
			IntervalSchedule: 5,
		}
		wc1 := ring.Subscribe(ctx, sub1)
		wc2 := ring.Subscribe(ctx, sub2)
		wc3 := ring.Subscribe(ctx, sub3)

		events := make([][]ringv2.Event, 3)
		var wg sync.WaitGroup
		wg.Add(3)

		for i, wc := range []<-chan ringv2.Event{wc1, wc2, wc3} {
			go func(wc <-chan ringv2.Event, i int) {
				for j := 0; j < 5; j++ {
					events[i] = append(events[i], <-wc)
				}
				wg.Done()
			}(wc, i)
		}

		wg.Wait()

		exp := []ringv2.Event{
			{Type: ringv2.EventTrigger, Values: []string{"mulder"}},
			{Type: ringv2.EventTrigger, Values: []string{"scully"}},
			{Type: ringv2.EventTrigger, Values: []string{"skinner"}},
			{Type: ringv2.EventTrigger, Values: []string{"mulder"}},
			{Type: ringv2.EventTrigger, Values: []string{"scully"}},
		}

		for i := range events {
			t.Run(fmt.Sprintf("client %d", i), func(t *testing.T) {
				if got, want := events[i], exp; !reflect.DeepEqual(got, want) {
					t.Fatalf("bad events: got %v, want %v", got, want)
				}
			})
		}
	})
}

// eventTest tests that for a given sequence of Add and Remove actions,
// a certain sequence of events is observed. eventTest will use the presence
// of EventAdd and EventRemove events to call ring.Add and ring.Remove.
//
// Given a sequence of want Events, eventTest will execute EventAdd and
// EventRemove, and attempt to observe EventTrigger.
func eventTest(t *testing.T, want []ringv2.Event) {
	t.Helper()
	t.Parallel()

	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		listener := pq.NewListener(dsn, time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Cleanup(func() {
			_ = listener.UnlistenAll()
			_ = listener.Close()
		})
		bus := NewBus(ctx, listener)
		ring, err := NewRing(db, bus, ringName(t.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = ring.Close() })

		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		sub := ringv2.Subscription{
			Name:             "test",
			Items:            1,
			IntervalSchedule: 5,
		}
		var wc <-chan ringv2.Event

		var got []ringv2.Event

		for _, event := range want {
			switch event.Type {
			case ringv2.EventAdd:
				entity := corev2.FixtureEntity(event.Values[0])
				namespace := corev3.FixtureNamespace(entity.Namespace)
				if err := namespaceStore.CreateOrUpdate(ctx, namespace); err != nil {
					t.Fatal(err)
				}
				if err := entityStore.UpdateEntity(ctx, entity); err != nil {
					t.Fatal(err)
				}
				if err := ring.Add(ctx, event.Values[0], 600); err != nil {
					t.Fatal(err)
				}
				got = append(got, event)
			case ringv2.EventRemove:
				if err := ring.Remove(ctx, event.Values[0]); err != nil {
					t.Fatal(err)
				}
				got = append(got, event)
			case ringv2.EventTrigger:
				if wc == nil {
					wc = ring.Subscribe(ctx, sub)
				}
				got = append(got, <-wc)
			}
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("bad events: got %v, want %v", got, want)
		}
	})
}

func TestRemoveNextTrigger(t *testing.T) {
	eventTest(t, []ringv2.Event{
		{Type: ringv2.EventAdd, Values: []string{"mulder"}},
		{Type: ringv2.EventAdd, Values: []string{"scully"}},
		{Type: ringv2.EventAdd, Values: []string{"skinner"}},
		{Type: ringv2.EventTrigger, Values: []string{"mulder"}},
		{Type: ringv2.EventTrigger, Values: []string{"scully"}},
		{Type: ringv2.EventRemove, Values: []string{"skinner"}},
		{Type: ringv2.EventTrigger, Values: []string{"mulder"}},
	})
}

func TestWatchAndAddAfter(t *testing.T) {
	eventTest(t, []ringv2.Event{
		{Type: ringv2.EventAdd, Values: []string{"byers"}},
		{Type: ringv2.EventAdd, Values: []string{"frohike"}},
		{Type: ringv2.EventTrigger, Values: []string{"byers"}},
		{Type: ringv2.EventAdd, Values: []string{"langly"}},
		{Type: ringv2.EventTrigger, Values: []string{"frohike"}},
		{Type: ringv2.EventTrigger, Values: []string{"langly"}},
		{Type: ringv2.EventTrigger, Values: []string{"byers"}},
	})
}

func TestWatchAfterAdd(t *testing.T) {
	t.Parallel()

	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		listener := pq.NewListener(dsn, time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Cleanup(func() {
			_ = listener.UnlistenAll()
			_ = listener.Close()
		})
		bus := NewBus(ctx, listener)
		ring, err := NewRing(db, bus, ringName(t.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = ring.Close()
		})

		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		entityName := "fowley"
		entity := corev2.FixtureEntity(entityName)
		namespace := corev3.FixtureNamespace(entity.Namespace)
		if err := namespaceStore.CreateOrUpdate(ctx, namespace); err != nil {
			t.Fatal(err)
		}
		if err := entityStore.UpdateEntity(ctx, entity); err != nil {
			t.Fatal(err)
		}
		if err := ring.Add(ctx, entityName, 600); err != nil {
			t.Fatal(err)
		}

		sub := ringv2.Subscription{
			Name:             "test",
			Items:            1,
			IntervalSchedule: 5,
		}
		wc := ring.Subscribe(ctx, sub)

		got := <-wc
		want := ringv2.Event{Type: ringv2.EventTrigger, Values: []string{entityName}}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("bad event: got %v, want %v", got, want)
		}
	})
}

func TestMultipleItems(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		listener := pq.NewListener(dsn, time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Cleanup(func() {
			_ = listener.UnlistenAll()
			_ = listener.Close()
		})
		bus := NewBus(ctx, listener)
		ring, err := NewRing(db, bus, ringName(t.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = ring.Close()
		})

		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		sub := ringv2.Subscription{
			Name:             "test",
			Items:            3,
			IntervalSchedule: 5,
		}

		items := []string{"byers", "frohike", "mulder", "scully", "skinner"}

		for _, item := range items {
			entity := corev2.FixtureEntity(item)
			namespace := corev3.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.CreateOrUpdate(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			if err := ring.Add(ctx, item, 600); err != nil {
				t.Fatal(err)
			}
		}

		wc := ring.Subscribe(ctx, sub)
		event := <-wc

		if got, want := event.Values, []string{"byers", "frohike", "mulder"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("bad values: got %v, want %v", got, want)
		}

		event = <-wc

		if got, want := event.Values, []string{"scully", "skinner", "byers"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("bad values: got %v, want %v", got, want)
		}

		event = <-wc

		// We wrapped around
		if got, want := event.Values, []string{"frohike", "mulder", "scully"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("bad values: got %v, want %v", got, want)
		}
	})
}

func TestRequestGreaterThanLenItems(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		listener := pq.NewListener(dsn, time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Cleanup(func() {
			_ = listener.UnlistenAll()
			_ = listener.Close()
		})
		bus := NewBus(ctx, listener)
		ring, err := NewRing(db, bus, ringName(t.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = ring.Close()
		})

		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		sub := ringv2.Subscription{
			Name:             "test",
			Items:            3,
			IntervalSchedule: 5,
		}

		items := []string{"byers", "frohike"}

		for _, item := range items {
			entity := corev2.FixtureEntity(item)
			namespace := corev3.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.CreateOrUpdate(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			if err := ring.Add(ctx, item, 600); err != nil {
				t.Fatal(err)
			}
		}

		wc := ring.Subscribe(ctx, sub)
		event := <-wc

		if got, want := event.Values, []string{"byers", "frohike", "byers"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("bad values: got %v, want %v", got, want)
		}

		event = <-wc

		if got, want := event.Values, []string{"frohike", "byers", "frohike"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("bad values: got %v, want %v", got, want)
		}

		event = <-wc

		// We wrapped around
		if got, want := event.Values, []string{"byers", "frohike", "byers"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("bad values: got %v, want %v", got, want)
		}
	})
}

func TestChannelNameTooLong(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		listener := pq.NewListener(dsn, time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Cleanup(func() {
			_ = listener.UnlistenAll()
			_ = listener.Close()
		})
		bus := NewBus(ctx, listener)
		ring, err := NewRing(db, bus, ringName(t.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = ring.Close()
		})

		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		sub := ringv2.Subscription{
			Name:             "extremely long test that has way too many characters in it, seriously, way too many",
			Items:            3,
			IntervalSchedule: 5,
		}

		items := []string{"byers", "frohike"}

		for _, item := range items {
			entity := corev2.FixtureEntity(item)
			namespace := corev3.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.CreateOrUpdate(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			if err := ring.Add(ctx, item, 600); err != nil {
				t.Fatal(err)
			}
		}

		wc := ring.Subscribe(ctx, sub)
		<-wc
		<-wc
	})
}

func TestListenChannelName(t *testing.T) {
	if got, want := ListenChannelName("foo", "bar"), "foo/bar"; got != want {
		t.Errorf("bad listen channel name: got %q, want %q", got, want)
	}
	if got := ListenChannelName("foo", "alsdkfjlasdkjflkasjdlfkjalskdjflkasjdlfkjasldkjflaksdjflkjasdlfkjlasdkfjlaskdjflkasjdlfkjasldkfjlasdk"); len(got) > 63 {
		t.Errorf("channel name too long: got length %d, want less than 63", len(got))
	}
}
