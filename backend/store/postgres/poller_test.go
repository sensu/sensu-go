package postgres

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func TestEntityConfigPoller(t *testing.T) {
	withPostgres(t, func(ctx context.Context, pool *pgxpool.Pool, dsn string) {
		estore := NewEntityConfigStore(pool)
		nsstore := NewNamespaceStore(pool)
		if err := nsstore.CreateIfNotExists(ctx, corev3.FixtureNamespace("default")); err != nil {
			t.Fatal(err)
		}
		if iState, err := estore.List(ctx, "", &store.SelectionPredicate{}); err != nil {
			t.Fatal(err)
		} else if len(iState) > 0 {
			t.Fatalf("unexpected non-empty entities: %v", iState)
		}

		watcherUnderTest := NewWatcher(estore, time.Millisecond*10, time.Millisecond*1500)

		// watch updates
		watchCtx, watchCancel := context.WithCancel(ctx)
		observedState := make(chan []*corev3.EntityConfig)
		go func() {
			req := storev2.ResourceRequest{
				APIVersion: "core/v3",
				Type:       "EntityConfig",
				StoreName:  "entity_configs",
			}
			watchEvents := watcherUnderTest.Watch(watchCtx, req)
			state := map[string]*corev3.EntityConfig{}
			for {
				select {
				case events := <-watchEvents:
					for _, event := range events {
						var ec corev3.EntityConfig
						if err := event.Value.UnwrapInto(&ec); err != nil {
							t.Error(err)
						}
						switch event.Type {
						case storev2.WatchCreate:
							if _, ok := state[ec.Metadata.Name]; ok {
								t.Errorf("unexpected double create: %v", event)
							}
							state[ec.Metadata.Name] = &ec
						case storev2.WatchUpdate:
							if _, ok := state[ec.Metadata.Name]; !ok {
								t.Errorf("unexpected update before create: %v", event)
							}
							state[ec.Metadata.Name] = &ec
						case storev2.WatchDelete:
							if _, ok := state[ec.Metadata.Name]; !ok {
								t.Errorf("unexpected delete: %v", event)
							}
							delete(state, ec.Metadata.Name)
						}
					}
				case <-watchCtx.Done():
					finalState := make([]*corev3.EntityConfig, 0, len(state))
					for _, e := range state {
						finalState = append(finalState, e)
					}
					observedState <- finalState
					return
				}
			}
		}()
		// wait for watcher to start up
		time.Sleep(time.Millisecond * 10)

		// generate some traffic
		for i := 0; i < 1000; i++ {
			ec := corev3.FixtureEntityConfig(fmt.Sprint(i))
			if err := estore.CreateIfNotExists(ctx, ec); err != nil {
				t.Fatal(err)
			}
		}
		for i := 0; i < 1000; i++ {
			if i%2 == 0 {
				continue
			}
			ec := corev3.FixtureEntityConfig(fmt.Sprint(i))
			ec.Metadata.Labels["foo"] = "bar"
			if err := estore.UpdateIfExists(ctx, ec); err != nil {
				t.Fatal(err)
			}
		}
		for i := 0; i < 1000; i++ {
			if i%10 == 0 {
				continue
			}
			if err := estore.Delete(ctx, "default", fmt.Sprint(i)); err != nil {
				t.Fatal(err)
			}
		}
		time.Sleep(time.Millisecond * 100)
		watchCancel()

		actualState := <-observedState

		sort.Slice(actualState, func(i, j int) bool {
			return actualState[i].Metadata.Name < actualState[j].Metadata.Name
		})
		expectedState, err := estore.List(ctx, "", &store.SelectionPredicate{})
		if err != nil {
			t.Fatal(err)
		}
		sort.Slice(expectedState, func(i, j int) bool {
			return expectedState[i].Metadata.Name < expectedState[j].Metadata.Name
		})
		if !cmp.Equal(expectedState, actualState) {
			t.Errorf("expected database and local state with watcher to match: %v", cmp.Diff(expectedState, actualState))
		}
	})
}
