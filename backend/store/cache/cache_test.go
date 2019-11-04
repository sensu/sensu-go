package cache

import (
	"context"
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/types"

	"github.com/coreos/etcd/integration"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/fixture"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixtureEntity(namespace, name string) *corev2.Entity {
	entity := corev2.FixtureEntity(name)
	entity.Namespace = namespace
	return entity
}

func TestCacheGet(t *testing.T) {
	cache := Resource{
		cache: buildCache([]corev2.Resource{
			fixtureEntity("a", "1"),
			fixtureEntity("a", "2"),
			fixtureEntity("a", "3"),
			fixtureEntity("a", "4"),
			fixtureEntity("a", "5"),
			fixtureEntity("a", "6"),
			fixtureEntity("b", "1"),
			fixtureEntity("b", "2"),
			fixtureEntity("b", "3"),
			fixtureEntity("b", "4"),
			fixtureEntity("b", "5"),
			fixtureEntity("b", "6"),
			fixtureEntity("c", "1"),
			fixtureEntity("c", "2"),
			fixtureEntity("c", "3"),
			fixtureEntity("c", "4"),
			fixtureEntity("c", "5"),
			fixtureEntity("c", "6"),
		},
			true,
		),
	}
	want := []Value{
		{Resource: fixtureEntity("b", "1"), Synth: dynamic.Synthesize(fixtureEntity("b", "1"))},
		{Resource: fixtureEntity("b", "2"), Synth: dynamic.Synthesize(fixtureEntity("b", "2"))},
		{Resource: fixtureEntity("b", "3"), Synth: dynamic.Synthesize(fixtureEntity("b", "3"))},
		{Resource: fixtureEntity("b", "4"), Synth: dynamic.Synthesize(fixtureEntity("b", "4"))},
		{Resource: fixtureEntity("b", "5"), Synth: dynamic.Synthesize(fixtureEntity("b", "5"))},
		{Resource: fixtureEntity("b", "6"), Synth: dynamic.Synthesize(fixtureEntity("b", "6"))},
	}
	got := cache.Get("b")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad resources: got %v, want %v", got, want)
	}
}

func TestCacheGetAll(t *testing.T) {
	cache := Resource{
		cache: buildCache([]corev2.Resource{
			fixtureEntity("a", "1"),
			fixtureEntity("a", "2"),
			fixtureEntity("b", "1"),
			fixtureEntity("b", "2"),
			fixtureEntity("b", "3"),
			fixtureEntity("c", "1"),
			fixtureEntity("c", "2"),
			fixtureEntity("c", "3"),
			fixtureEntity("c", "4"),
		},
			false,
		),
	}
	got := cache.GetAll()
	assert.Equal(t, 9, len(got))
	want := []Value{
		{Resource: fixtureEntity("a", "1")},
		{Resource: fixtureEntity("a", "2")},
		{Resource: fixtureEntity("b", "1")},
		{Resource: fixtureEntity("b", "2")},
		{Resource: fixtureEntity("b", "3")},
		{Resource: fixtureEntity("c", "1")},
		{Resource: fixtureEntity("c", "2")},
		{Resource: fixtureEntity("c", "3")},
		{Resource: fixtureEntity("c", "4")},
	}
	for _, v := range want {
		assert.Contains(t, got, v)
	}
}

func TestBuildCache(t *testing.T) {
	resource1 := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "resource1", Namespace: "default"}}
	resource2 := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "resource2", Namespace: "default"}}
	resource3 := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "resource3", Namespace: "acme"}}

	cache := buildCache([]corev2.Resource{resource1, resource2, resource3}, false)

	assert.Len(t, cache["acme"], 1)
	assert.Len(t, cache["default"], 2)
	assert.Len(t, cache, 2)
}

func TestResourceUpdateCache(t *testing.T) {
	cacher := Resource{
		cache: make(map[string][]Value),
	}
	resource0 := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "resource0", Namespace: "default"}, Foo: "bar"}
	resource1 := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "resource1", Namespace: "default"}, Foo: "bar"}
	resource0Bis := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "resource0", Namespace: "default"}, Foo: "baz"}
	resource1Bis := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "resource1", Namespace: "default"}, Foo: "qux"}

	// Add a resource
	cacher.updates = append(cacher.updates, store.WatchEventResource{
		Resource: resource1,
		Action:   store.WatchCreate,
	})
	cacher.updateCache(context.Background())
	assert.Len(t, cacher.cache["default"], 1)

	// Add a second resource. It should be alphabetically sorted and therefore at
	// the beginning of the namespace cache values even if it was appended at the
	// end
	cacher.updates = append(cacher.updates, store.WatchEventResource{
		Resource: resource0, Action: store.WatchCreate,
	})
	cacher.updateCache(context.Background())
	assert.Len(t, cacher.cache["default"], 2)
	assert.Equal(t, resource0, cacher.cache["default"][0].Resource)
	assert.Equal(t, resource1, cacher.cache["default"][1].Resource)

	// Update the resources
	updates := []store.WatchEventResource{
		store.WatchEventResource{Resource: resource0Bis, Action: store.WatchUpdate},
		store.WatchEventResource{Resource: resource1Bis, Action: store.WatchUpdate},
	}
	cacher.updates = append(cacher.updates, updates...)
	cacher.updateCache(context.Background())
	assert.Len(t, cacher.cache["default"], 2)
	assert.Equal(t, resource0Bis, cacher.cache["default"][0].Resource.(*fixture.Resource))
	assert.Equal(t, resource1Bis, cacher.cache["default"][1].Resource.(*fixture.Resource))

	// Delete the resources
	deletes := []store.WatchEventResource{
		store.WatchEventResource{Resource: resource1Bis, Action: store.WatchDelete},
		store.WatchEventResource{Resource: resource0Bis, Action: store.WatchDelete},
	}
	cacher.updates = append(cacher.updates, deletes...)
	cacher.updateCache(context.Background())
	assert.Len(t, cacher.cache["default"], 0)

	// Invalid watch event
	var nilResource *fixture.Resource
	cacher.updates = append(cacher.updates, store.WatchEventResource{
		Resource: nilResource,
		Action:   store.WatchCreate,
	})
	cacher.updateCache(context.Background())
	assert.Len(t, cacher.cache["default"], 0)
}

func TestResourceRebuild(t *testing.T) {
	c := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer c.Terminate(t)
	client := c.RandClient()
	s := etcd.NewStore(client, "store")
	require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("default")))
	ctx := store.NamespaceContext(context.Background(), "default")

	cacher := Resource{
		cache:     make(map[string][]Value),
		client:    client,
		resourceT: &fixture.Resource{},
	}

	// Empty store
	cacher.updates = append(cacher.updates, store.WatchEventResource{
		Action: store.WatchError,
	})
	cacher.updateCache(ctx)
	assert.Len(t, cacher.cache["default"], 0)

	// Resource added to a new namespace
	foo := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "foo", Namespace: "default"}}
	if err := s.CreateOrUpdateResource(ctx, foo); err != nil {
		t.Fatal(err)
	}
	cacher.updates = append(cacher.updates, store.WatchEventResource{
		Action: store.WatchError,
	})
	cacher.updateCache(ctx)
	assert.Len(t, cacher.cache["default"], 1)

	// Resource added to an existing namespace
	bar := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "bar", Namespace: "default"}}
	if err := s.CreateOrUpdateResource(ctx, bar); err != nil {
		t.Fatal(err)
	}
	cacher.updates = append(cacher.updates, store.WatchEventResource{
		Action: store.WatchError,
	})
	cacher.updateCache(ctx)
	assert.Len(t, cacher.cache["default"], 2)

	// Resource updated
	bar.Foo = "acme"
	if err := s.CreateOrUpdateResource(ctx, bar); err != nil {
		t.Fatal(err)
	}
	cacher.updates = append(cacher.updates, store.WatchEventResource{
		Action: store.WatchError,
	})
	cacher.updateCache(ctx)
	assert.Len(t, cacher.cache["default"], 2)

	// Resource deleted
	if err := s.DeleteResource(ctx, bar.StorePrefix(), bar.GetObjectMeta().Name); err != nil {
		t.Fatal(err)
	}
	cacher.updates = append(cacher.updates, store.WatchEventResource{
		Action: store.WatchError,
	})
	cacher.updateCache(ctx)
	assert.Len(t, cacher.cache["default"], 1)
}
