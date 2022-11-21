package v2

import (
	"reflect"
	"testing"

	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/stretchr/testify/assert"
)

func fixtureEntity(namespace, name string) *corev3.EntityConfig {
	entity := corev3.FixtureEntityConfig(name)
	entity.Metadata.Namespace = namespace
	return entity
}

func TestCacheGet(t *testing.T) {
	cache := Resource[*corev3.EntityConfig, corev3.EntityConfig]{
		cache: buildCache([]*corev3.EntityConfig{
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
	want := []Value[*corev3.EntityConfig, corev3.EntityConfig]{
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
	cache := Resource[*corev3.EntityConfig, corev3.EntityConfig]{
		cache: buildCache([]*corev3.EntityConfig{
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
	want := []Value[*corev3.EntityConfig, corev3.EntityConfig]{
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

	cache := buildCache([]*corev3.EntityConfig{resource1, resource2, resource3}, false)

	assert.Len(t, cache["acme"], 1)
	assert.Len(t, cache["default"], 2)
	assert.Len(t, cache, 2)
}
