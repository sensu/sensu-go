package v2

import (
	"context"
	"fmt"
	"testing"

	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

var mockTime = crock.NewTime(time.Unix(0, 0))

func init() {
	mockTime.Resolution = time.Millisecond
	mockTime.Multiplier = 100
	time.TimeProxy = mockTime
}

func TestEntityCacheIntegration(t *testing.T) {
	ctx := store.NamespaceContext(context.Background(), "default")
	mockTime.Start()
	defer mockTime.Stop()
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	store := etcdstore.NewStore(client)

	// Add namespace resource
	namespace := corev2.FixtureNamespace("default")
	req := storev2.NewResourceRequestFromV2Resource(namespace)
	wrapper, err := wrap.V2Resource(namespace)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateOrUpdate(ctx, req, wrapper); err != nil {
		t.Fatal(err)
	}

	fixtures := []Value{}

	// Populate store with some initial entities
	for i := 0; i < 9; i++ {
		fixture := corev3.FixtureEntityConfig(fmt.Sprintf("%d", i))
		fixture.Metadata.Name = fmt.Sprintf("%d", i)
		fixture.EntityClass = corev2.EntityProxyClass
		fixtures = append(fixtures, getCacheValue(fixture, true))
		req = storev2.NewResourceRequestFromResource(fixture)
		wrapper, err := storev2.WrapResource(fixture)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
	}

	otherFixtures := []Value{}

	// Include some entities from a non-default namespace
	namespace = corev2.FixtureNamespace("other")
	req = storev2.NewResourceRequestFromV2Resource(namespace)
	wrapper, err = wrap.V2Resource(namespace)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateOrUpdate(ctx, req, wrapper); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		fixture := corev3.FixtureEntityConfig(fmt.Sprintf("%d", i))
		fixture.Metadata.Name = fmt.Sprintf("%d", i)
		fixture.Metadata.Namespace = "other"
		fixture.EntityClass = corev2.EntityProxyClass
		otherFixtures = append(otherFixtures, getCacheValue(fixture, true))
		req = storev2.NewResourceRequestFromResource(fixture)
		wrapper, err := storev2.WrapResource(fixture)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
	}

	cacheCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache, err := New(cacheCtx, store, &corev3.EntityConfig{}, true)
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

	newEntity := corev3.FixtureEntityConfig("new")
	newEntity.EntityClass = corev2.EntityProxyClass
	req = storev2.NewResourceRequestFromResource(newEntity)
	entWrapper, err := storev2.WrapResource(newEntity)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateOrUpdate(ctx, req, entWrapper); err != nil {
		t.Fatal(err)
	}
	<-watcher

	got := cache.Get("default")

	if got, want := got[len(got)-1], getCacheValue(newEntity, true); got.Resource.GetMetadata().Name != want.Resource.GetMetadata().Name {
		t.Errorf("bad entity: got %s, want %s", got.Resource.GetMetadata().Name, want.Resource.GetMetadata().Name)
	}

	req = storev2.NewResourceRequestFromResource(newEntity)
	if err := store.Delete(ctx, req); err != nil {
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
		if got[i].Resource.GetMetadata().Namespace != want[i].Resource.GetMetadata().Namespace {
			return false
		}
		if got[i].Resource.GetMetadata().Name != want[i].Resource.GetMetadata().Name {
			return false
		}
	}
	return true
}
