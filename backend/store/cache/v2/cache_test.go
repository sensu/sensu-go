package v2

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/types/dynamic"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/tests/v3/integration"
)

func fixtureEntity(namespace, name string) *corev3.EntityConfig {
	entity := corev3.FixtureEntityConfig(name)
	entity.Metadata.Namespace = namespace
	return entity
}

func TestCacheGet(t *testing.T) {
	cache := Resource{
		cache: buildCache([]corev3.Resource{
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
		cache: buildCache([]corev3.Resource{
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
	resource1 := corev3.FixtureEntityConfig("resource1")
	resource2 := corev3.FixtureEntityConfig("resource2")
	resource3 := corev3.FixtureEntityConfig("resource3")
	resource3.Metadata.Namespace = "acme"

	cache := buildCache([]corev3.Resource{resource1, resource2, resource3}, false)

	assert.Len(t, cache["acme"], 1)
	assert.Len(t, cache["default"], 2)
	assert.Len(t, cache, 2)
}

func TestResourceRebuild(t *testing.T) {
	ctx := store.NamespaceContext(context.Background(), "default")
	integration.BeforeTestExternal(t)
	c := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer c.Terminate(t)
	client := c.RandClient()
	store := etcdstore.NewStore(client)

	// Add namespace resource
	namespace := corev2.FixtureNamespace("default")
	req := storev2.NewResourceRequestFromV2Resource(ctx, namespace)
	wrapper, err := wrap.V2Resource(namespace)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateOrUpdate(req, wrapper); err != nil {
		t.Fatal(err)
	}

	cacher := Resource{
		cache:     make(map[string][]Value),
		client:    client,
		resourceT: &corev3.EntityConfig{},
	}

	// Resource added to a new namespace
	foo := corev3.FixtureEntityConfig("foo")
	req = storev2.NewResourceRequestFromResource(ctx, foo)
	fooWrapper, err := storev2.WrapResource(foo)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateOrUpdate(req, fooWrapper); err != nil {
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
	bar := corev3.FixtureEntityConfig("bar")
	req = storev2.NewResourceRequestFromResource(ctx, bar)
	barWrapper, err := storev2.WrapResource(bar)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateOrUpdate(req, barWrapper); err != nil {
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
	bar.User = "acme"
	req = storev2.NewResourceRequestFromResource(ctx, bar)
	barWrapper, err = storev2.WrapResource(bar)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateOrUpdate(req, barWrapper); err != nil {
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
	req = storev2.NewResourceRequestFromResource(ctx, bar)
	if err := store.Delete(req); err != nil {
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
