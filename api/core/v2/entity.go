package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	utilstrings "github.com/sensu/sensu-go/util/strings"
)

const (
	// EntityAgentClass is the name of the class given to agent entities.
	EntityAgentClass = "agent"

	// EntityProxyClass is the name of the class given to proxy entities.
	EntityProxyClass = "proxy"

	// EntityBackendClass is the name of the class given to backend entities.
	EntityBackendClass = "backend"

	// Redacted is filled in for fields that contain sensitive information
	Redacted = "REDACTED"
)

// DefaultRedactFields contains the default fields to redact
var DefaultRedactFields = []string{"password", "passwd", "pass", "api_key",
	"api_token", "access_key", "secret_key", "private_key", "secret"}

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

// NewEntity creates a new Entity.
func NewEntity(meta ObjectMeta) *Entity {
	return &Entity{ObjectMeta: meta}
}

func redactMap(m map[string]string, redact []string) map[string]string {
	if len(redact) == 0 {
		redact = DefaultRedactFields
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		if utilstrings.FoundInArray(k, redact) {
			result[k] = Redacted
		} else {
			result[k] = v
		}
	}
	return result
}

// GetRedactedEntity redacts the entity according to the entity's Redact fields.
// A redacted copy is returned. The copy contains pointers to the original's
// memory, with different Labels and Annotations.
func (e *Entity) GetRedactedEntity() *Entity {
	if e == nil {
		return nil
	}
	if e.Labels == nil && e.Annotations == nil {
		return e
	}
	ent := &Entity{}
	*ent = *e
	ent.Annotations = redactMap(e.Annotations, e.Redact)
	ent.Labels = redactMap(e.Labels, e.Redact)
	return ent
}

// MarshalJSON implements the json.Marshaler interface.
func (e *Entity) MarshalJSON() ([]byte, error) {
	// Redact the entity before marshalling the entity so we don't leak any
	// sensitive information
	e = e.GetRedactedEntity()

	type Clone Entity
	clone := (*Clone)(e)

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
		System: System{
			Arch: "amd64",
		},
	}
}

// URIPath returns the path component of a Entity URI.
func (e *Entity) URIPath() string {
	return fmt.Sprintf("/api/core/v2/namespaces/%s/entities/%s", url.PathEscape(e.Namespace), url.PathEscape(e.Name))
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

// EntityFields returns a set of fields that represent that resource
func EntityFields(r Resource) map[string]string {
	resource := r.(*Entity)
	return map[string]string{
		"entity.name":          resource.ObjectMeta.Name,
		"entity.namespace":     resource.ObjectMeta.Namespace,
		"entity.deregister":    strconv.FormatBool(resource.Deregister),
		"entity.entity_class":  resource.EntityClass,
		"entity.subscriptions": strings.Join(resource.Subscriptions, ","),
	}
}
