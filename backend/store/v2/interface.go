package v2

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"

	"github.com/mitchellh/hashstructure"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
)

// Interface provides access to the various stores that Sensu supports.
type Interface interface {
	ConfigStoreGetter
	EntityConfigStoreGetter
	EntityStateStoreGetter
	NamespaceStoreGetter
	EventStoreGetter
	EntityStoreGetter
	SilencesStoreGetter
}

// Wrapper is an abstraction of a store wrapper.
type Wrapper interface {
	Unwrap() (corev3.Resource, error)
	UnwrapInto(interface{}) error
}

// WrapList is a list of Wrappers.
type WrapList interface {
	Unwrap() ([]corev3.Resource, error)
	UnwrapInto(interface{}) error
	Len() int
}

// ETag represents a unique hash of the resource.
type ETag []byte

// String returns the base64-encoded unquoted form of the ETag.
func (e ETag) String() string {
	return base64.RawStdEncoding.EncodeToString(e)
}

// DecodeETag attempts to parse an unquoted encoded ETag. It returns an error if the
// input is not base64-encoded.
func DecodeETag(data string) (ETag, error) {
	b, err := base64.RawStdEncoding.DecodeString(data)
	return ETag(b), err
}

func (e ETag) Equals(other ETag) bool {
	return bytes.Equal(e, other)
}

// ETagFromStruct is for types that don't want to compute etag in the database
func ETagFromStruct(v interface{}) ETag {
	hash, err := hashstructure.Hash(v, nil)
	if err != nil {
		panic(err)
	}
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, hash)
	return ETag(b)
}

// EntityConfigStoreGetter gets you an EntityConfigStore.
type EntityConfigStoreGetter interface {
	GetEntityConfigStore() EntityConfigStore
}

// EntityStateStoreGetter gets you an EntityStateStore.
type EntityStateStoreGetter interface {
	GetEntityStateStore() EntityStateStore
}

// NamespaceStoreGetter gets you a NamespaceStore.
type NamespaceStoreGetter interface {
	GetNamespaceStore() NamespaceStore
}

// ConfigStoreGetter gets you a ConfigStore.
type ConfigStoreGetter interface {
	GetConfigStore() ConfigStore
}

// EventStoreGetter gets you an EventStore.
type EventStoreGetter interface {
	GetEventStore() store.EventStore
}

// EntityStoreGetter gets you an EntityStore.
type EntityStoreGetter interface {
	GetEntityStore() store.EntityStore
}

// SilencesStoreGetter gets you a SilencesStore
type SilencesStoreGetter interface {
	GetSilencesStore() SilencesStore
}

// ConfigStore specifies the interface of a v2 store.
type ConfigStore interface {
	// CreateOrUpdate creates or updates the wrapped resource.
	CreateOrUpdate(context.Context, ResourceRequest, Wrapper) error

	// UpdateIfExists updates the resource with the wrapped resource, but only
	// if it already exists in the store.
	UpdateIfExists(context.Context, ResourceRequest, Wrapper) error

	// CreateIfNotExists writes the wrapped resource to the store, but only if
	// it does not already exist.
	CreateIfNotExists(context.Context, ResourceRequest, Wrapper) error

	// Get gets a wrapped resource from the store.
	Get(context.Context, ResourceRequest) (Wrapper, error)

	// Delete deletes a resource from the store.
	Delete(context.Context, ResourceRequest) error

	// List lists all resources specified by the resource request, and the
	// selection predicate.
	List(context.Context, ResourceRequest, *store.SelectionPredicate) (WrapList, error)

	// Count returns a count of all resources matching the set of constraints
	// specified by the resource request.
	Count(context.Context, ResourceRequest) (int, error)

	// Exists returns true if the resource indicated by the request exists
	Exists(context.Context, ResourceRequest) (bool, error)

	// Patch patches the resource
	Patch(context.Context, ResourceRequest, patch.Patcher) error

	// Watch provides a channel for receiving updates to a particular resource
	// or resource collection
	Watch(context.Context, ResourceRequest) <-chan []WatchEvent

	// Initialize
	Initialize(context.Context, InitializeFunc) error
}

// NamespaceStore provides an interface for interacting with namespaces.
type NamespaceStore interface {
	// CreateOrUpdate creates or updates a corev3.Namespace resource.
	CreateOrUpdate(context.Context, *corev3.Namespace) error

	// UpdateIfExists updates the corev3.Namespace resource, but only if it
	// already exists in the store.
	UpdateIfExists(context.Context, *corev3.Namespace) error

	// CreateIfNotExists writes the corev3.Namespace resource to the store, but
	// only if it does not already exist.
	CreateIfNotExists(context.Context, *corev3.Namespace) error

	// Get gets a corev3.Namespace from the store.
	Get(context.Context, string) (*corev3.Namespace, error)

	// Delete deletes the corev3.Namespace, from the store, corresponding with
	// the given name. It will return an error of the namespace is not empty.
	Delete(context.Context, string) error

	// List lists all corev3.Namespace resources.
	List(context.Context, *store.SelectionPredicate) ([]*corev3.Namespace, error)

	// Count returns a count of all namespaces
	Count(context.Context) (int, error)

	// Exists returns true if the corev3.Namespace with the provided name exists.
	Exists(context.Context, string) (bool, error)

	// Patch patches the corev3.Namespace resource with the provided name.
	Patch(context.Context, string, patch.Patcher) error

	// IsEmpty returns whether the corev3.Namespace with the provided name
	// is empty or if it contains other resources.
	IsEmpty(context.Context, string) (bool, error)
}

// EntityConfigStore provides an interface for interacting with entity configs.
type EntityConfigStore interface {
	// CreateOrUpdate creates or updates a corev3.EntityConfig resource.
	CreateOrUpdate(context.Context, *corev3.EntityConfig) error

	// UpdateIfExists updates the corev3.EntityConfig resource, but only if it
	// already exists in the store.
	UpdateIfExists(context.Context, *corev3.EntityConfig) error

	// CreateIfNotExists writes the corev3.EntityConfig resource to the store, but
	// only if it does not already exist.
	CreateIfNotExists(context.Context, *corev3.EntityConfig) error

	// Get gets a corev3.EntityConfig from the store.
	Get(context.Context, string, string) (*corev3.EntityConfig, error)

	// Delete deletes the corev3.EntityConfig, from the store, corresponding
	// with the given namespace and name.
	Delete(context.Context, string, string) error

	// List lists all corev3.EntityConfig resources.
	List(context.Context, string, *store.SelectionPredicate) ([]*corev3.EntityConfig, error)

	// Count returns a count of entities by namespace and optionally by entity
	// class
	Count(ctx context.Context, namespace, entityClass string) (int, error)

	// Exists returns true if the corev3.EntityConfig exists for the provided
	// namespace and name.
	Exists(context.Context, string, string) (bool, error)

	// Patch patches the corev3.EntityConfig resource with the provided
	// namespace and name.
	Patch(context.Context, string, string, patch.Patcher) error

	// Watch creates a watcher for the key space defined by the namespace and name
	// given. There are three possible modes. If namespace and name are blank, then
	// the watcher watches all EntityConfigs in the system. If the namespace is
	// supplied but the name is blank, the watcher watches all EntityConfigs in a
	// particular namespace. If the namespace and name are both supplied, the watcher
	// watches a single EntityConfig.
	Watch(c context.Context, namespace, name string) <-chan []WatchEvent
}

// EntityStateStore provides an interface for interacting with entity states.
type EntityStateStore interface {
	// CreateOrUpdate creates or updates a corev3.EntityState resource.
	CreateOrUpdate(context.Context, *corev3.EntityState) error

	// UpdateIfExists updates the corev3.EntityState resource, but only if it
	// already exists in the store.
	UpdateIfExists(context.Context, *corev3.EntityState) error

	// CreateIfNotExists writes the corev3.EntityState resource to the store, but
	// only if it does not already exist.
	CreateIfNotExists(context.Context, *corev3.EntityState) error

	// Get gets a corev3.EntityState from the store.
	Get(context.Context, string, string) (*corev3.EntityState, error)

	// Delete deletes the corev3.EntityState, from the store, corresponding
	// with the given namespace and name.
	Delete(context.Context, string, string) error

	// List lists all corev3.EntityState resources.
	List(context.Context, string, *store.SelectionPredicate) ([]*corev3.EntityState, error)

	// Count returns a count of entity states by namespace
	Count(context.Context, string) (int, error)

	// Exists returns true if the corev3.EntityState exists for the provided
	// namespace and name.
	Exists(context.Context, string, string) (bool, error)

	// Patch patches the corev3.EntityState resource with the provided
	// namespace and name.
	Patch(context.Context, string, string, patch.Patcher) error
}

// KeepaliveStore provides an interface for interacting with keepalives.
type KeepaliveStore interface {
	// DeleteFailingKeepalive deletes a failing keepalive record for a given entity.
	DeleteFailingKeepalive(ctx context.Context, entityConfig *corev3.EntityConfig) error

	// GetFailingKeepalives returns a slice of failing keepalives.
	// TODO: create a corev3.KeepaliveRecord type so we can remove the
	// dependency on corev2
	GetFailingKeepalives(ctx context.Context) ([]*corev2.KeepaliveRecord, error)

	// UpdateFailingKeepalive updates the given entity keepalive with the given expiration
	// in unix timestamp format
	UpdateFailingKeepalive(ctx context.Context, entityConfig *corev3.EntityConfig, expiration int64) error
}

// Initialize sets up a cluster with the default resources & config.
type InitializeFunc func(context.Context, Interface) error

type SilencesStore interface {
	// GetSilences gets a list of silences belonging to the provided namespace.
	GetSilences(ctx context.Context, namespace string) ([]*corev2.Silenced, error)

	// GetSilencesByCheck gets a list of silences belonging to the provided namespace and check.
	GetSilencesByCheck(ctx context.Context, namespace, check string) ([]*corev2.Silenced, error)

	// GetSilencesBySubscription gets a list of silences belonging to the provided namespace and matching
	// the subscriptions provided.
	GetSilencesBySubscription(ctx context.Context, namespace string, subscriptions []string) ([]*corev2.Silenced, error)

	// GetSilenceByName gets an individual silence by name
	GetSilenceByName(ctx context.Context, namespace, name string) (*corev2.Silenced, error)

	// UpdateSilence updates a silence
	UpdateSilence(ctx context.Context, si *corev2.Silenced) error

	// GetSilencesByName gets a list of silences matching the provided names
	GetSilencesByName(ctx context.Context, namespace string, names []string) ([]*corev2.Silenced, error)

	// DeleteSilences deletes one or more named silences
	DeleteSilences(ctx context.Context, namespace string, names []string) error
}
