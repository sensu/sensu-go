package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	utilstrings "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
)

const (
	// EntitiesResource is the name of this resource type
	EntitiesResource = "entities"

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

// StorePrefix returns the path prefix to this resource in the store
func (e *Entity) StorePrefix() string {
	return EntitiesResource
}

// URIPath returns the path component of an entity URI.
func (e *Entity) URIPath() string {
	if e.Namespace == "" {
		return path.Join(URLPrefix, EntitiesResource, url.PathEscape(e.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(e.Namespace), EntitiesResource, url.PathEscape(e.Name))
}

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
		ObjectMeta: ObjectMeta{
			Namespace: "default",
			Name:      name,
		},
		EntityClass:   "host",
		User:          "agent1",
		Subscriptions: []string{"linux", GetEntitySubscription(name)},
		Redact: []string{
			"password",
		},
		System: System{
			Arch:           "amd64",
			OS:             "linux",
			Platform:       "Gentoo",
			PlatformFamily: "lol",
			Network: Network{
				Interfaces: []NetworkInterface{
					{
						Name: "eth0",
						MAC:  "return of the",
						Addresses: []string{
							"127.0.0.1",
						},
					},
				},
			},
			LibCType:      "glibc",
			VMSystem:      "kvm",
			VMRole:        "host",
			CloudProvider: "aws",
			FloatType:     "hard",
			Processes: []*Process{
				{
					Name: "sensu-agent",
				},
			},
		},
		LastSeen:          12345,
		SensuAgentVersion: "0.0.1",
	}
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

// SetNamespace sets the namespace of the resource.
func (e *Entity) SetNamespace(namespace string) {
	e.Namespace = namespace
}

// SetObjectMeta sets the meta of the resource.
func (e *Entity) SetObjectMeta(meta ObjectMeta) {
	e.ObjectMeta = meta
}

func (e *Entity) RBACName() string {
	return "entities"
}

// SetName sets the name of the resource.
func (e *Entity) SetName(name string) {
	e.Name = name
}
