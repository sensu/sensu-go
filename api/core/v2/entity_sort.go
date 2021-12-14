package v2

import (
	"fmt"
	"sort"
)

const (
	EntitySortName     = "NAME"
	EntitySortLastSeen = "LAST_SEEN"
)

func SortEntities(entities []*Entity, ordering string, desc bool) {
	if len(ordering) == 0 {
		return
	}

	var sortIf sort.Interface
	switch ordering {
	case EntitySortName:
		sortIf = SortEntitiesByID(entities, !desc)
	case EntitySortLastSeen:
		sortIf = SortEntitiesByLastSeen(entities, !desc)
	default:
		panic(fmt.Sprintf("SortEntities: unknown order: %s", ordering))
	}

	sort.Sort(sortIf)
}

// SortEntitiesByPredicate can be used to sort a given collection using a given
// predicate.
func SortEntitiesByPredicate(es []*Entity, fn func(a, b *Entity) bool) sort.Interface {
	return &entitySorter{entities: es, byFn: fn}
}

// SortEntitiesByID can be used to sort a given collection of entities by their
// IDs.
func SortEntitiesByID(es []*Entity, asc bool) sort.Interface {
	if asc {
		return SortEntitiesByPredicate(es, func(a, b *Entity) bool {
			return a.Name < b.Name
		})
	}
	return SortEntitiesByPredicate(es, func(a, b *Entity) bool {
		return a.Name > b.Name
	})
}

// SortEntitiesByLastSeen can be used to sort a given collection of entities by
// last time each was seen.
func SortEntitiesByLastSeen(es []*Entity, asc bool) sort.Interface {
	return SortEntitiesByPredicate(es, func(a, b *Entity) bool {
		if asc {
			return a.LastSeen < b.LastSeen
		}
		return a.LastSeen > b.LastSeen
	})
}

type entitySorter struct {
	entities []*Entity
	byFn     func(a, b *Entity) bool
}

// Len implements sort.Interface.
func (s *entitySorter) Len() int {
	return len(s.entities)
}

// Swap implements sort.Interface.
func (s *entitySorter) Swap(i, j int) {
	s.entities[i], s.entities[j] = s.entities[j], s.entities[i]
}

// Less implements sort.Interface.
func (s *entitySorter) Less(i, j int) bool {
	return s.byFn(s.entities[i], s.entities[j])
}
