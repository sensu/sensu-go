// +build integration

package schedulerd

import (
	"context"
	"fmt"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
	store "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/types"
)

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
	fixtures := []*corev2.Entity{}

	// Populate store with some initial entities
	for i := 0; i < 9; i++ {
		fixture := corev2.FixtureEntity(fmt.Sprintf("%d", i))
		fixture.Name = fmt.Sprintf("%d", i)
		fixture.EntityClass = corev2.EntityProxyClass
		fixtures = append(fixtures, fixture)
		if err := store.UpdateEntity(ctx, fixture); err != nil {
			t.Fatal(err)
		}
	}

	otherFixtures := []*corev2.Entity{}

	// Include some entities from a non-default namespace
	if err := store.CreateNamespace(context.Background(), &corev2.Namespace{Name: "other"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		fixture := corev2.FixtureEntity(fmt.Sprintf("%d", i))
		fixture.Name = fmt.Sprintf("%d", i)
		fixture.Namespace = "other"
		fixture.EntityClass = corev2.EntityProxyClass
		otherFixtures = append(otherFixtures, fixture)
		if err := store.UpdateEntity(ctx, fixture); err != nil {
			t.Fatal(err)
		}
	}

	cacheCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache, err := NewEntityCache(cacheCtx, store)
	if err != nil {
		t.Fatal(err)
	}

	watcher := cache.Watch(cacheCtx)

	if got, want := cache.GetEntities("default"), fixtures; !checkEntities(got, want) {
		t.Fatalf("bad entities")
	}

	if got, want := cache.GetEntities("notdefault"), []*corev2.Entity{}; !checkEntities(got, want) {
		t.Fatal("bad entities")
	}

	if got, want := cache.GetEntities("other"), otherFixtures; !checkEntities(got, want) {
		t.Fatal("bad entities")
	}

	newEntity := corev2.FixtureEntity("new")
	newEntity.EntityClass = corev2.EntityProxyClass
	if err := store.UpdateEntity(ctx, newEntity); err != nil {
		t.Fatal(err)
	}

	<-watcher

	got := cache.GetEntities("default")

	if got, want := got[len(got)-1], newEntity; got.Name != want.Name {
		t.Errorf("bad entity: got %s, want %s", got.Name, want.Name)
	}

	if err := store.DeleteEntity(ctx, newEntity); err != nil {
		t.Fatal(err)
	}

	<-watcher

	if got, want := cache.GetEntities("default"), fixtures; !checkEntities(got, want) {
		t.Errorf("bad entities")
	}

}

func checkEntities(got, want []*corev2.Entity) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Namespace != want[i].Namespace {
			return false
		}
		if got[i].Name != want[i].Name {
			return false
		}
	}
	return true
}
