package cache

import (
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types/dynamic"
)

func fixtureEntity(namespace, name string) *corev2.Entity {
	entity := corev2.FixtureEntity(name)
	entity.Namespace = namespace
	return entity
}

func TestCacheGet(t *testing.T) {
	cache := Resource{
		sliceCache: []Value{
			{Resource: fixtureEntity("a", "1"), Synth: dynamic.Synthesize(fixtureEntity("a", "1"))},
			{Resource: fixtureEntity("a", "2"), Synth: dynamic.Synthesize(fixtureEntity("a", "2"))},
			{Resource: fixtureEntity("a", "3"), Synth: dynamic.Synthesize(fixtureEntity("a", "3"))},
			{Resource: fixtureEntity("a", "4"), Synth: dynamic.Synthesize(fixtureEntity("a", "4"))},
			{Resource: fixtureEntity("a", "5"), Synth: dynamic.Synthesize(fixtureEntity("a", "5"))},
			{Resource: fixtureEntity("a", "6"), Synth: dynamic.Synthesize(fixtureEntity("a", "6"))},
			{Resource: fixtureEntity("b", "1"), Synth: dynamic.Synthesize(fixtureEntity("b", "1"))},
			{Resource: fixtureEntity("b", "2"), Synth: dynamic.Synthesize(fixtureEntity("b", "2"))},
			{Resource: fixtureEntity("b", "3"), Synth: dynamic.Synthesize(fixtureEntity("b", "3"))},
			{Resource: fixtureEntity("b", "4"), Synth: dynamic.Synthesize(fixtureEntity("b", "4"))},
			{Resource: fixtureEntity("b", "5"), Synth: dynamic.Synthesize(fixtureEntity("b", "5"))},
			{Resource: fixtureEntity("b", "6"), Synth: dynamic.Synthesize(fixtureEntity("b", "6"))},
			{Resource: fixtureEntity("c", "1"), Synth: dynamic.Synthesize(fixtureEntity("c", "1"))},
			{Resource: fixtureEntity("c", "2"), Synth: dynamic.Synthesize(fixtureEntity("c", "2"))},
			{Resource: fixtureEntity("c", "3"), Synth: dynamic.Synthesize(fixtureEntity("c", "3"))},
			{Resource: fixtureEntity("c", "4"), Synth: dynamic.Synthesize(fixtureEntity("c", "4"))},
			{Resource: fixtureEntity("c", "5"), Synth: dynamic.Synthesize(fixtureEntity("c", "5"))},
			{Resource: fixtureEntity("c", "6"), Synth: dynamic.Synthesize(fixtureEntity("c", "6"))},
		},
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
