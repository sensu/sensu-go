package types

import (
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
)

// Validate returns an error if the entity is invalid.
func (e *Entity) Validate() error {
	if err := ValidateName(e.ID); err != nil {
		return errors.New("entity id " + err.Error())
	}

	if err := ValidateName(e.Class); err != nil {
		return errors.New("entity class " + err.Error())
	}

	if e.Environment == "" {
		return errors.New("environment must be set")
	}

	if e.Organization == "" {
		return errors.New("organization must be set")
	}

	return nil
}

// Get implements govaluate.Parameters
func (e *Entity) Get(name string) (interface{}, error) {
	return dynamic.GetField(e, name)
}

// SetExtendedAttributes sets the serialized ExtendedAttributes of the entity.
func (e *Entity) SetExtendedAttributes(b []byte) {
	e.ExtendedAttributes = b
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Entity) UnmarshalJSON(b []byte) error {
	return dynamic.Unmarshal(b, e)
}

// MarshalJSON implements the json.Marshaler interface.
func (e *Entity) MarshalJSON() ([]byte, error) {
	// Redact the entity before marshalling the entity so we don't leak any
	// sensitive information
	redactedEntity, err := dynamic.Redact(e, e.Redact...)
	if err != nil {
		return nil, err
	}

	return dynamic.Marshal(redactedEntity)
}

// GetEntitySubscription returns the entity subscription, using the format
// "entity:entityID"
func GetEntitySubscription(entityID string) string {
	return fmt.Sprintf("entity:%s", entityID)
}

// FixtureEntity returns a testing fixture for an Entity object.
func FixtureEntity(id string) *Entity {
	return &Entity{
		ID:               id,
		Class:            "host",
		Subscriptions:    []string{"linux"},
		Environment:      "default",
		Organization:     "default",
		KeepaliveTimeout: 120,
	}
}

// URIPath returns the path component of a Entity URI.
func (e *Entity) URIPath() string {
	return fmt.Sprintf("/entities/%s", url.PathEscape(e.ID))
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
			return a.ID < b.ID
		})
	}
	return SortEntitiesByPredicate(es, func(a, b *Entity) bool {
		return a.ID > b.ID
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
