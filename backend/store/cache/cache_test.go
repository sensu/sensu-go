package cache

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/backend/store/etcd"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/dynamic"
	"github.com/sensu/sensu-go/testing/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/tests/v3/integration"
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

func TestResourceRebuild(t *testing.T) {
	integration.BeforeTestExternal(t)
	c := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer c.Terminate(t)
	client := c.RandClient()
	s := etcd.NewStore(client, "store")
	require.NoError(t, s.CreateNamespace(context.Background(), corev2.FixtureNamespace("default")))
	ctx := store.NamespaceContext(context.Background(), "default")

	cacher := Resource{
		cache:     make(map[string][]Value),
		client:    client,
		resourceT: &fixture.Resource{},
	}

	// Resource added to a new namespace
	foo := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "foo", Namespace: "default"}}
	if err := s.CreateOrUpdateResource(ctx, foo); err != nil {
		t.Fatal(err)
	}
	if updates, err := cacher.rebuild(ctx); err != nil {
		t.Fatal(err)
	} else if !updates {
		t.Fatal("expected updates")
	}
	assert.Len(t, cacher.cache["default"], 1)
	assert.Equal(t, int64(1), cacher.Count())

	// Resource added to an existing namespace
	bar := &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "bar", Namespace: "default"}}
	if err := s.CreateOrUpdateResource(ctx, bar); err != nil {
		t.Fatal(err)
	}
	if updates, err := cacher.rebuild(ctx); err != nil {
		t.Fatal(err)
	} else if !updates {
		t.Fatal("expected updates")
	}
	assert.Len(t, cacher.cache["default"], 2)
	assert.Equal(t, int64(2), cacher.Count())

	// Resource updated
	bar.Foo = "acme"
	if err := s.CreateOrUpdateResource(ctx, bar); err != nil {
		t.Fatal(err)
	}
	if updates, err := cacher.rebuild(ctx); err != nil {
		t.Fatal(err)
	} else if !updates {
		t.Fatal("expected updates")
	}
	assert.Len(t, cacher.cache["default"], 2)
	assert.Equal(t, int64(2), cacher.Count())

	// Resource deleted
	if err := s.DeleteResource(ctx, bar.StorePrefix(), bar.GetObjectMeta().Name); err != nil {
		t.Fatal(err)
	}
	if updates, err := cacher.rebuild(ctx); err != nil {
		t.Fatal(err)
	} else if !updates {
		t.Fatal("expected updates")
	}
	assert.Len(t, cacher.cache["default"], 1)
	assert.Equal(t, int64(1), cacher.Count())
}

func nonNamespacedCache(count int) Resource {
	resources := []corev2.Resource{}
	for i := 0; i < count; i++ {
		resources = append(resources, &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: fmt.Sprintf("namespace-%d", i)}})
	}
	return Resource{
		cache: buildCache(resources, false),
	}
}

func BenchmarkGetEmpty(b *testing.B) {
	cache := nonNamespacedCache(20)
	for n := 0; n < b.N; n++ {
		values := cache.Get("")
		for range values {
		}
	}
}

func BenchmarkGetAll(b *testing.B) {
	cache := nonNamespacedCache(20)
	for n := 0; n < b.N; n++ {
		values := cache.GetAll()
		for range values {
		}
	}
}
