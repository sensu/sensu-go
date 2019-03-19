package schedulerd

import (
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func fixtureEntity(namespace, name string) *corev2.Entity {
	entity := corev2.FixtureEntity(name)
	entity.Namespace = namespace
	return entity
}

func TestEntityCacheGetEntities(t *testing.T) {
	cache := EntityCache{
		sliceCache: []*corev2.Entity{
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
	}
	want := []*corev2.Entity{
		fixtureEntity("b", "1"),
		fixtureEntity("b", "2"),
		fixtureEntity("b", "3"),
		fixtureEntity("b", "4"),
		fixtureEntity("b", "5"),
		fixtureEntity("b", "6"),
	}
	got := cache.GetEntities("b")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad entities: got %v, want %v", got, want)
	}
}
