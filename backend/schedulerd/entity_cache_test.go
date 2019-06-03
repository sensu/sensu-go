package schedulerd

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

func TestEntityCacheGetEntities(t *testing.T) {
	cache := EntityCache{
		sliceCache: []EntityCacheValue{
			{Entity: fixtureEntity("a", "1"), Synth: dynamic.Synthesize(fixtureEntity("a", "1"))},
			{Entity: fixtureEntity("a", "2"), Synth: dynamic.Synthesize(fixtureEntity("a", "2"))},
			{Entity: fixtureEntity("a", "3"), Synth: dynamic.Synthesize(fixtureEntity("a", "3"))},
			{Entity: fixtureEntity("a", "4"), Synth: dynamic.Synthesize(fixtureEntity("a", "4"))},
			{Entity: fixtureEntity("a", "5"), Synth: dynamic.Synthesize(fixtureEntity("a", "5"))},
			{Entity: fixtureEntity("a", "6"), Synth: dynamic.Synthesize(fixtureEntity("a", "6"))},
			{Entity: fixtureEntity("b", "1"), Synth: dynamic.Synthesize(fixtureEntity("b", "1"))},
			{Entity: fixtureEntity("b", "2"), Synth: dynamic.Synthesize(fixtureEntity("b", "2"))},
			{Entity: fixtureEntity("b", "3"), Synth: dynamic.Synthesize(fixtureEntity("b", "3"))},
			{Entity: fixtureEntity("b", "4"), Synth: dynamic.Synthesize(fixtureEntity("b", "4"))},
			{Entity: fixtureEntity("b", "5"), Synth: dynamic.Synthesize(fixtureEntity("b", "5"))},
			{Entity: fixtureEntity("b", "6"), Synth: dynamic.Synthesize(fixtureEntity("b", "6"))},
			{Entity: fixtureEntity("c", "1"), Synth: dynamic.Synthesize(fixtureEntity("c", "1"))},
			{Entity: fixtureEntity("c", "2"), Synth: dynamic.Synthesize(fixtureEntity("c", "2"))},
			{Entity: fixtureEntity("c", "3"), Synth: dynamic.Synthesize(fixtureEntity("c", "3"))},
			{Entity: fixtureEntity("c", "4"), Synth: dynamic.Synthesize(fixtureEntity("c", "4"))},
			{Entity: fixtureEntity("c", "5"), Synth: dynamic.Synthesize(fixtureEntity("c", "5"))},
			{Entity: fixtureEntity("c", "6"), Synth: dynamic.Synthesize(fixtureEntity("c", "6"))},
		},
	}
	want := []EntityCacheValue{
		{Entity: fixtureEntity("b", "1"), Synth: dynamic.Synthesize(fixtureEntity("b", "1"))},
		{Entity: fixtureEntity("b", "2"), Synth: dynamic.Synthesize(fixtureEntity("b", "2"))},
		{Entity: fixtureEntity("b", "3"), Synth: dynamic.Synthesize(fixtureEntity("b", "3"))},
		{Entity: fixtureEntity("b", "4"), Synth: dynamic.Synthesize(fixtureEntity("b", "4"))},
		{Entity: fixtureEntity("b", "5"), Synth: dynamic.Synthesize(fixtureEntity("b", "5"))},
		{Entity: fixtureEntity("b", "6"), Synth: dynamic.Synthesize(fixtureEntity("b", "6"))},
	}
	got := cache.GetEntities("b")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad entities: got %v, want %v", got, want)
	}
}
