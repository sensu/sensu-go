// +build integration

package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
	store "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/types"
)

var mockTime = crock.NewTime(time.Unix(0, 0))

func init() {
	mockTime.Resolution = time.Millisecond
	mockTime.Multiplier = 100
	time.TimeProxy = mockTime
}

func TestEntityCacheIntegration(t *testing.T) {
	mockTime.Start()
	defer mockTime.Stop()
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	store := store.NewStore(client, e.Name())

	if err := store.CreateNamespace(context.Background(), types.FixtureNamespace("default")); err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, "default")
	fixtures := []Value{}

	// Populate store with some initial entities
	for i := 0; i < 9; i++ {
		fixture := corev2.FixtureEntity(fmt.Sprintf("%d", i))
		fixture.Name = fmt.Sprintf("%d", i)
		fixture.EntityClass = corev2.EntityProxyClass
		fixtures = append(fixtures, getCacheValue(fixture, true))
		if err := store.UpdateEntity(ctx, fixture); err != nil {
			t.Fatal(err)
		}
	}

	otherFixtures := []Value{}

	// Include some entities from a non-default namespace
	if err := store.CreateNamespace(context.Background(), &corev2.Namespace{Name: "other"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		fixture := corev2.FixtureEntity(fmt.Sprintf("%d", i))
		fixture.Name = fmt.Sprintf("%d", i)
		fixture.Namespace = "other"
		fixture.EntityClass = corev2.EntityProxyClass
		otherFixtures = append(otherFixtures, getCacheValue(fixture, true))
		if err := store.UpdateEntity(ctx, fixture); err != nil {
			t.Fatal(err)
		}
	}

	cacheCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache, err := New(cacheCtx, client, &corev2.Entity{}, true)
	if err != nil {
		t.Fatal(err)
	}

	watcher := cache.Watch(cacheCtx)

	if got, want := cache.Get("default"), fixtures; !checkEntities(got, want) {
		t.Fatalf("bad entities")
	}

	if got, want := cache.Get("notdefault"), []Value{}; !checkEntities(got, want) {
		t.Fatal("bad entities")
	}

	if got, want := cache.Get("other"), otherFixtures; !checkEntities(got, want) {
		t.Fatal("bad entities")
	}

	newEntity := corev2.FixtureEntity("new")
	newEntity.EntityClass = corev2.EntityProxyClass
	if err := store.UpdateEntity(ctx, newEntity); err != nil {
		t.Fatal(err)
	}
	<-watcher

	got := cache.Get("default")

	if got, want := got[len(got)-1], getCacheValue(newEntity, true); got.Resource.GetObjectMeta().Name != want.Resource.GetObjectMeta().Name {
		t.Errorf("bad entity: got %s, want %s", got.Resource.GetObjectMeta().Name, want.Resource.GetObjectMeta().Name)
	}

	if err := store.DeleteEntity(ctx, newEntity); err != nil {
		t.Fatal(err)
	}

	<-watcher

	if got, want := cache.Get("default"), fixtures; !checkEntities(got, want) {
		t.Errorf("bad entities")
	}

}

func checkEntities(got, want []Value) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Resource.GetObjectMeta().Namespace != want[i].Resource.GetObjectMeta().Namespace {
			return false
		}
		if got[i].Resource.GetObjectMeta().Name != want[i].Resource.GetObjectMeta().Name {
			return false
		}
	}
	return true
}
