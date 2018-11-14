package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"

	"github.com/sensu/sensu-go/types/dynamic"
)

const (
	// EntityAgentClass is the name of the class given to agent entities.
	EntityAgentClass = "agent"

	// EntityProxyClass is the name of the class given to proxy entities.
	EntityProxyClass = "proxy"

	// EntityBackendClass is the name of the class given to backend entities.
	EntityBackendClass = "backend"
)

// Validate returns an error if the entity is invalid.
func (e *Entity) Validate() error {
	if err := ValidateName(e.Name); err != nil {
		return errors.New("entity name " + err.Error())
	}

	if err := ValidateName(e.EntityClass); err != nil {
		return errors.New("entity class " + err.Error())
	}

	if e.Namespace == "" {
		return errors.New("namespace must be set")
	}

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (e *Entity) MarshalJSON() ([]byte, error) {
	// Redact the entity before marshalling the entity so we don't leak any
	// sensitive information
	redactedEntity, err := dynamic.Redact(e, e.Redact...)
	if err != nil {
		return nil, err
	}

	type Clone Entity
	clone := (*Clone)(redactedEntity.(*Entity))

	return json.Marshal(clone)
}

// GetEntitySubscription returns the entity subscription, using the format
// "entity:entityName"
func GetEntitySubscription(entityName string) string {
	return fmt.Sprintf("entity:%s", entityName)
}

// FixtureEntity returns a testing fixture for an Entity object.
func FixtureEntity(name string) *Entity {
	return &Entity{
		EntityClass:   "host",
		Subscriptions: []string{"linux"},
		ObjectMeta: ObjectMeta{
			Namespace: "default",
			Name:      name,
		},
	}
}

// URIPath returns the path component of a Entity URI.
func (e *Entity) URIPath() string {
	return fmt.Sprintf("/entities/%s", url.PathEscape(e.Name))
}

//
// Sorting

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
func SortEntitiesByLastSeen(es []*Entity) sort.Interface {
	return SortEntitiesByPredicate(es, func(a, b *Entity) bool {
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
