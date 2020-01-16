package store

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

// ErrAlreadyExists is returned when an object already exists
type ErrAlreadyExists struct {
	Key string
}

func (e *ErrAlreadyExists) Error() string {
	return fmt.Sprintf("could not create the key %s", e.Key)
}

// ErrDecode is returned when an object could not be decoded
type ErrDecode struct {
	Key string
	Err error
}

func (e *ErrDecode) Error() string {
	return fmt.Sprintf("could not decode the key %s: %s", e.Key, e.Err.Error())
}

// ErrEncode is returned when an object could not be decoded
type ErrEncode struct {
	Key string
	Err error
}

func (e *ErrEncode) Error() string {
	return fmt.Sprintf("could not encode the key %s: %s", e.Key, e.Err.Error())
}

// ErrNamespaceMissing is returned when the user tries to manipulate a resource
// within a namespace that does not exist
type ErrNamespaceMissing struct {
	Namespace string
}

func (e *ErrNamespaceMissing) Error() string {
	return fmt.Sprintf("the namespace %s does not exist", e.Namespace)
}

// ErrNotFound is returned when a key is not found in the store
type ErrNotFound struct {
	Key string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("key %s not found", e.Key)
}

// ErrNotValid is returned when an object failed validation
type ErrNotValid struct {
	Err error
}

func (e *ErrNotValid) Error() string {
	return fmt.Sprintf("resource is invalid: %s", e.Err.Error())
}

// ErrInternal is returned when something generally bad happened while
// interacting with the store. Other, more specific errors should be
// returned when appropriate.
//
// The backend will use ErrInternal to detect if an error is unrecoverable.
// It should only be used to signal that the underlying database is not
// functional
type ErrInternal struct {
	Message string
}

func (e *ErrInternal) Error() string {
	return fmt.Sprintf("internal error: %s", e.Message)
}

// SelectionPredicate represents the way to select resources from storage
type SelectionPredicate struct {
	// Continue provides the key from which the selection should start. If
	// returned empty from the store, it indicates that there's no additional
	// resources available
	Continue string
	// Limit indicates the number of resources to retrieve
	Limit int64
	// Subcollection represents a sub-collection of the primary collection
	Subcollection string
}

// A WatchEventCheckConfig contains the modified store object and the action that occured
// during the modification.
type WatchEventCheckConfig struct {
	CheckConfig *types.CheckConfig
	Action      WatchActionType
}

// A WatchEventHookConfig contains the modified asset object and the action that occurred
// during the modification.
type WatchEventHookConfig struct {
	HookConfig *types.HookConfig
	Action     WatchActionType
}

// WatchEventTessenConfig is a notification that the tessen config store has been updated.
type WatchEventTessenConfig struct {
	TessenConfig *corev2.TessenConfig
	Action       WatchActionType
}

// WatchEventResource is a store event about a specific resource
type WatchEventResource struct {
	Resource corev2.Resource
	Action   WatchActionType
}

// Store is used to abstract the durable storage used by the Sensu backend
// processses. Each Sensu resources is represented by its own interface. A
// MockStore is available in order to mock a store implementation
type Store interface {
	// AssetStore provides an interface for managing checks assets
	AssetStore

	// AuthenticationStore provides an interface for managing the JWT secret
	AuthenticationStore

	// CheckConfigStore provides an interface for managing checks configuration
	CheckConfigStore

	// ClusterIDStore provides an interface for managing the sensu cluster id
	ClusterIDStore

	// EntityStore provides an interface for managing entities
	EntityStore

	// EventStore provides an interface for managing events
	EventStore

	// EventFilterStore provides an interface for managing events filters
	EventFilterStore

	// HandlerStore provides an interface for managing events handlers
	HandlerStore

	// HealthStore provides an interface for getting cluster health information
	HealthStore

	// HookConfigStore provides an interface for managing hooks configuration
	HookConfigStore

	// KeepaliveStore provides an interface for managing entities keepalives
	KeepaliveStore

	// MutatorStore provides an interface for managing events mutators
	MutatorStore

	// NamespaceStore provides an interface for managing namespaces
	NamespaceStore

	// ClusterRoleStore provides an interface for managing cluster roles
	ClusterRoleStore

	// ClusterRoleBindingStore provides an interface for managing cluster role bindings
	ClusterRoleBindingStore

	// RoleStore provides an interface for managing roles
	RoleStore

	// RoleBindingStore provides an interface for managing role bindings
	RoleBindingStore

	// SilencedStore provides an interface for managing silenced entries,
	// consisting of entities, subscriptions and/or checks
	SilencedStore

	// TessenConfigStore provides an interface for managing the tessen configuration
	TessenConfigStore

	// UserStore provides an interface for managing users
	UserStore

	// ExtensionRegistry tracks third-party extensions.
	ExtensionRegistry

	// ResourceStore ...
	ResourceStore

	// NewInitializer returns the Initializer interfaces, which provides the
	// required mechanism to verify if a store is initialized
	NewInitializer() (Initializer, error)
}

// AssetStore provides methods for managing checks assets
type AssetStore interface {
	// DeleteAssetByName deletes an asset using the given name and the
	// namespace stored in ctx.
	DeleteAssetByName(ctx context.Context, name string) error

	// GetAssets returns all assets in the given ctx's namespace. A nil
	// slice with no error is returned if none were found.
	GetAssets(ctx context.Context, pred *SelectionPredicate) ([]*types.Asset, error)

	// GetAssetByName returns an asset using the given name and the namespace
	// stored in ctx. The resulting asset is nil if none was found.
	GetAssetByName(ctx context.Context, name string) (*types.Asset, error)

	// UpdateAsset creates or updates a given asset.
	UpdateAsset(ctx context.Context, asset *types.Asset) error
}

// AuthenticationStore provides methods for managing the JWT secret
type AuthenticationStore interface {
	// CreateJWTSecret create the given JWT secret and returns an error if it was
	// unsuccessful or if a secret already exists.
	CreateJWTSecret(secret []byte) error

	// GetJWTSecret returns the JWT secret.
	GetJWTSecret() ([]byte, error)

	// UpdateJWTSecret updates the JWT secret with the given secret.
	UpdateJWTSecret(secret []byte) error
}

// CheckConfigStore provides methods for managing checks configuration
type CheckConfigStore interface {
	// DeleteCheckConfigByName deletes a check's configuration using the given name
	// and the namespace stored in ctx.
	DeleteCheckConfigByName(ctx context.Context, name string) error

	// GetCheckConfigs returns all checks configurations in the given ctx's
	// namespace. A nil slice with no error is returned if none
	// were found.
	GetCheckConfigs(ctx context.Context, pred *SelectionPredicate) ([]*types.CheckConfig, error)

	// GetCheckConfigByName returns a check's configuration using the given name
	// and the namespace stored in ctx. The resulting check is
	// nil if none was found.
	GetCheckConfigByName(ctx context.Context, name string) (*types.CheckConfig, error)

	// UpdateCheckConfig creates or updates a given check's configuration.
	UpdateCheckConfig(ctx context.Context, check *types.CheckConfig) error

	// GetCheckConfigWatcher returns a channel that emits CheckConfigWatchEvents notifying
	// the caller that a CheckConfig was updated. If the watcher runs into a terminal error
	// or the context passed is cancelled, then the channel will be closed. The caller must
	// restart the watcher, if needed.
	GetCheckConfigWatcher(ctx context.Context) <-chan WatchEventCheckConfig
}

// ClusterIDStore provides methods for managing the sensu cluster id
type ClusterIDStore interface {
	// CreateClusterID creates a sensu cluster id
	CreateClusterID(context.Context, string) error

	// GetClusterID gets the sensu cluster id
	GetClusterID(context.Context) (string, error)
}

// ClusterRoleBindingStore provides methods for managing RBAC cluster role
// bindings
type ClusterRoleBindingStore interface {
	// Create a given cluster role binding
	CreateClusterRoleBinding(ctx context.Context, clusterRoleBinding *types.ClusterRoleBinding) error

	// CreateOrUpdateRole overwrites the given cluster role binding
	CreateOrUpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *types.ClusterRoleBinding) error

	// DeleteRole deletes a cluster role binding using the given name.
	DeleteClusterRoleBinding(ctx context.Context, name string) error

	// GetRole returns a cluster role binding using the given name. An error is
	// returned if no binding was found
	GetClusterRoleBinding(ctx context.Context, name string) (*types.ClusterRoleBinding, error)

	// ListRoles returns all cluster role binding. An error is returned if no
	// binding were found
	ListClusterRoleBindings(ctx context.Context, pred *SelectionPredicate) (clusterRoleBindings []*types.ClusterRoleBinding, err error)

	// UpdateRole creates or updates a given cluster role binding.
	UpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *types.ClusterRoleBinding) error
}

// ClusterRoleStore provides methods for managing RBAC cluster roles and rules
type ClusterRoleStore interface {
	// Create a given cluster role
	CreateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error

	// CreateOrUpdateClusterRole overwrites the given cluster role
	CreateOrUpdateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error

	// DeleteClusterRole deletes a cluster role using the given name.
	DeleteClusterRole(ctx context.Context, name string) error

	// GetClusterRole returns a cluster role using the given name. An error is
	// returned if no role was found
	GetClusterRole(ctx context.Context, name string) (*types.ClusterRole, error)

	// ListClusterRoles returns all cluster roles. An error is returned if no
	// roles were found
	ListClusterRoles(ctx context.Context, pred *SelectionPredicate) (clusterRoles []*types.ClusterRole, err error)

	// UpdateClusterRole creates or updates a given cluster role.
	UpdateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error
}

// HookConfigStore provides methods for managing hooks configuration
type HookConfigStore interface {
	// DeleteHookConfigByName deletes a hook's configuration using the given name
	// and the namespace stored in ctx.
	DeleteHookConfigByName(ctx context.Context, name string) error

	// GetHookConfigs returns all hooks configurations in the given ctx's
	// namespace. A nil slice with no error is returned if none
	// were found.
	GetHookConfigs(ctx context.Context, pred *SelectionPredicate) ([]*types.HookConfig, error)

	// GetHookConfigByName returns a hook's configuration using the given name and
	// the namespace stored in ctx. The resulting hook is nil if none was found.
	GetHookConfigByName(ctx context.Context, name string) (*types.HookConfig, error)

	// UpdateHookConfig creates or updates a given hook's configuration.
	UpdateHookConfig(ctx context.Context, check *types.HookConfig) error
}

// EntityStore provides methods for managing entities
type EntityStore interface {
	// DeleteEntity deletes an entity using the given entity struct.
	DeleteEntity(ctx context.Context, entity *types.Entity) error

	// DeleteEntityByName deletes an entity using the given name and the
	// namespace stored in ctx.
	DeleteEntityByName(ctx context.Context, name string) error

	// GetEntities returns all entities in the given ctx's namespace. A nil slice
	// with no error is returned if none were found.
	GetEntities(ctx context.Context, pred *SelectionPredicate) ([]*types.Entity, error)

	// GetEntityByName returns an entity using the given name and the namespace stored
	// in ctx. The resulting entity is nil if none was found.
	GetEntityByName(ctx context.Context, name string) (*types.Entity, error)

	// UpdateEntity creates or updates a given entity.
	UpdateEntity(ctx context.Context, entity *types.Entity) error
}

// EventStore provides methods for managing events
type EventStore interface {
	// DeleteEventByEntityCheck deletes an event using the given entity and check,
	// within the namespace stored in ctx.
	DeleteEventByEntityCheck(ctx context.Context, entity, check string) error

	// GetEvents returns all events in the given ctx's namespace. A nil slice with
	// no error is returned if none were found.
	GetEvents(ctx context.Context, pred *SelectionPredicate) ([]*corev2.Event, error)

	// GetEventsByEntity returns all events for the given entity within the ctx's
	// namespace. A nil slice with no error is returned if none were found.
	GetEventsByEntity(ctx context.Context, entity string, pred *SelectionPredicate) ([]*corev2.Event, error)

	// GetEventByEntityCheck returns an event using the given entity and check,
	// within the namespace stored in ctx. The resulting event
	// is nil if none was found.
	GetEventByEntityCheck(ctx context.Context, entity, check string) (*types.Event, error)

	// UpdateEvent creates or updates a given event. It returns the updated
	// event, which may be the same as the event that was passed in, and the
	// previous event, if one existed, as well as any error that occurred.
	UpdateEvent(ctx context.Context, event *types.Event) (old, new *types.Event, err error)
}

// EventFilterStore provides methods for managing events filters
type EventFilterStore interface {
	// DeleteEventFilterByName deletes an event filter using the given name and the
	// namespace stored in ctx.
	DeleteEventFilterByName(ctx context.Context, name string) error

	// GetEventFilters returns all filters in the given ctx's namespace. A nil
	// slice with no error is returned if none were found.
	GetEventFilters(ctx context.Context, pred *SelectionPredicate) ([]*types.EventFilter, error)

	// GetEventFilterByName returns a filter using the given name and the
	// namespace stored in ctx. The resulting filter is nil if none was found.
	GetEventFilterByName(ctx context.Context, name string) (*types.EventFilter, error)

	// UpdateEventFilter creates or updates a given filter.
	UpdateEventFilter(ctx context.Context, filter *types.EventFilter) error
}

// HandlerStore provides methods for managing events handlers
type HandlerStore interface {
	// DeleteHandlerByName deletes a handler using the given name and the
	// namespace stored in ctx.
	DeleteHandlerByName(ctx context.Context, name string) error

	// GetHandlers returns all handlers in the given ctx's namespace. A nil slice
	// with no error is returned if none were found.
	GetHandlers(ctx context.Context, pred *SelectionPredicate) ([]*types.Handler, error)

	// GetHandlerByName returns a handler using the given name and the namespace
	// stored in ctx. The resulting handler is nil if none was found.
	GetHandlerByName(ctx context.Context, name string) (*types.Handler, error)

	// UpdateHandler creates or updates a given handler.
	UpdateHandler(ctx context.Context, handler *types.Handler) error
}

// HealthStore provides methods for cluster health
type HealthStore interface {
	GetClusterHealth(ctx context.Context, cluster clientv3.Cluster, etcdClientTLSConfig *tls.Config) *types.HealthResponse
}

// KeepaliveStore provides methods for managing entities keepalives
type KeepaliveStore interface {
	// DeleteFailingKeepalive deletes a failing keepalive record for a given entity.
	DeleteFailingKeepalive(ctx context.Context, entity *types.Entity) error

	// GetFailingKeepalives returns a slice of failing keepalives.
	GetFailingKeepalives(ctx context.Context) ([]*types.KeepaliveRecord, error)

	// UpdateFailingKeepalive updates the given entity keepalive with the given expiration
	// in unix timestamp format
	UpdateFailingKeepalive(ctx context.Context, entity *types.Entity, expiration int64) error
}

// MutatorStore provides methods for managing events mutators
type MutatorStore interface {
	// DeleteMutatorByName deletes a mutator using the given name and the
	// namespace stored in ctx.
	DeleteMutatorByName(ctx context.Context, name string) error

	// GetMutators returns all mutators in the given ctx's namespace. A nil slice
	// with no error is returned if none were found.
	GetMutators(ctx context.Context, pred *SelectionPredicate) ([]*types.Mutator, error)

	// GetMutatorByName returns a mutator using the given name and the
	// namespace stored in ctx. The resulting mutator is nil if
	// none was found.
	GetMutatorByName(ctx context.Context, name string) (*types.Mutator, error)

	// UpdateMutator creates or updates a given mutator.
	UpdateMutator(ctx context.Context, mutator *types.Mutator) error
}

// NamespaceStore provides methods for managing namespaces
type NamespaceStore interface {
	// CreateNamespace creates a given namespace
	CreateNamespace(ctx context.Context, namespace *types.Namespace) error

	// DeleteNamespace deletes a namespace using the given name.
	DeleteNamespace(ctx context.Context, name string) error

	// ListNamespaces returns all namespaces. A nil slice with no error is
	// returned if none were found.
	ListNamespaces(ctx context.Context, pred *SelectionPredicate) ([]*types.Namespace, error)

	// GetNamespace returns a namespace using the given name. The
	// result is nil if none was found.
	GetNamespace(ctx context.Context, name string) (*types.Namespace, error)

	// UpdateNamespace updates an existing namespace.
	UpdateNamespace(ctx context.Context, org *types.Namespace) error
}

// ResourceStore ...
type ResourceStore interface {
	CreateResource(ctx context.Context, resource corev2.Resource) error

	CreateOrUpdateResource(ctx context.Context, resource corev2.Resource) error

	DeleteResource(ctx context.Context, kind, name string) error

	GetResource(ctx context.Context, name string, resource corev2.Resource) error

	ListResources(ctx context.Context, kind string, resources interface{}, pred *SelectionPredicate) error
}

// RoleBindingStore provides methods for managing RBAC role bindings
type RoleBindingStore interface {
	// Create a given role binding
	CreateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error

	// CreateOrUpdateRole overwrites the given role binding
	CreateOrUpdateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error

	// DeleteRole deletes a role binding using the given name.
	DeleteRoleBinding(ctx context.Context, name string) error

	// GetRole returns a role binding using the given name. An error is returned
	// if no binding was found
	GetRoleBinding(ctx context.Context, name string) (*types.RoleBinding, error)

	// ListRoles returns all role binding. An error is returned if no binding were
	// found
	ListRoleBindings(ctx context.Context, pred *SelectionPredicate) (roleBindings []*types.RoleBinding, err error)

	// UpdateRole creates or updates a given role binding.
	UpdateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error
}

// RoleStore provides methods for managing RBAC roles and rules
type RoleStore interface {
	// Create a given role
	CreateRole(ctx context.Context, role *types.Role) error

	// CreateOrUpdateRole overwrites the given role
	CreateOrUpdateRole(ctx context.Context, role *types.Role) error

	// DeleteRole deletes a role using the given name.
	DeleteRole(ctx context.Context, name string) error

	// GetRole returns a role using the given name. An error is returned if no
	// role was found
	GetRole(ctx context.Context, name string) (*types.Role, error)

	// ListRoles returns all roles. An error is returned if no roles were found
	ListRoles(ctx context.Context, pred *SelectionPredicate) (roles []*types.Role, err error)

	// UpdateRole creates or updates a given role.
	UpdateRole(ctx context.Context, role *types.Role) error
}

// SilencedStore provides methods for managing silenced entries,
// consisting of entities, subscriptions and/or checks
type SilencedStore interface {
	// DeleteSilencedEntryByName deletes an entry using the given id.
	DeleteSilencedEntryByName(ctx context.Context, id ...string) error

	// GetSilencedEntries returns all entries. A nil slice with no error is
	// returned if none were found.
	GetSilencedEntries(ctx context.Context) ([]*types.Silenced, error)

	// GetSilencedEntriesByCheckName returns all entries for the given check
	// within the ctx's namespace. A nil slice with no error is
	// returned if none were found.
	GetSilencedEntriesByCheckName(ctx context.Context, check string) ([]*types.Silenced, error)

	// GetSilencedEntriesByCheckName returns all entries for the given subscription
	// within the ctx's namespace. A nil slice with no error is
	// returned if none were found.
	GetSilencedEntriesBySubscription(ctx context.Context, subscriptions ...string) ([]*types.Silenced, error)

	// GetSilencedEntryByName returns an entry using the given id and the
	// namespace stored in ctx. The resulting entry is nil if
	// none was found.
	GetSilencedEntryByName(ctx context.Context, id string) (*types.Silenced, error)

	// UpdateHandler creates or updates a given entry.
	UpdateSilencedEntry(ctx context.Context, entry *types.Silenced) error

	// GetSilencedEntriesByName gets all the named silenced entries.
	GetSilencedEntriesByName(ctx context.Context, id ...string) ([]*types.Silenced, error)
}

// TessenConfigStore provides methods for managing the Tessen configuration
type TessenConfigStore interface {
	// CreateOrUpdateTessenConfig creates or updates the tessen configuration
	CreateOrUpdateTessenConfig(context.Context, *corev2.TessenConfig) error

	// GetTessenConfig gets the tessen configuration
	GetTessenConfig(context.Context) (*corev2.TessenConfig, error)

	// GetTessenConfigWatcher returns a tessen config watcher
	GetTessenConfigWatcher(context.Context) <-chan WatchEventTessenConfig
}

// UserStore provides methods for managing users
type UserStore interface {
	// AuthenticateUser attempts to authenticate a user with the given username
	// and hashed password. An error is returned if the user does not exist, is
	// disabled or the given password does not match.
	AuthenticateUser(ctx context.Context, username, password string) (*types.User, error)

	// CreateUsern creates a new user with the given user struct.
	CreateUser(user *types.User) error

	// DeleteUser deletes a user using the given user struct.
	DeleteUser(ctx context.Context, user *types.User) error

	// GetUser returns a user using the given username.
	GetUser(ctx context.Context, username string) (*types.User, error)

	// GetUsers returns all enabled users. A nil slice with no error is
	// returned if none were found.
	GetUsers() ([]*types.User, error)

	// GetUsers returns all users, including the disabled ones. A nil slice with
	// no error is  returned if none were found.
	GetAllUsers(pred *SelectionPredicate) ([]*types.User, error)

	// UpdateHandler updates a given user.
	UpdateUser(user *types.User) error
}

// Initializer provides methods to verify if a store is initialized
type Initializer interface {
	// Close closes the session to the store and unlock any mutex
	Close() error

	// FlagAsInitialized marks the store as initialized
	FlagAsInitialized() error

	// IsInitialized returns a boolean error that indicates if the store has been
	// initialized or not
	IsInitialized() (bool, error)

	// Lock locks a mutex to avoid competing writes
	Lock() error
}

// ErrNoExtension is returned when a named extension does not exist.
var ErrNoExtension = errors.New("the extension does not exist")

// ExtensionRegistry registers and tracks Sensu extensions.
type ExtensionRegistry interface {
	// RegisterExtension registers an extension. It associates an extension type and
	// name with a URL. The registry assumes that the extension provides
	// a handler and a mutator named 'name'.
	RegisterExtension(context.Context, *types.Extension) error

	// DeregisterExtension deregisters an extension. If the extension does not exist,
	// nil error is returned.
	DeregisterExtension(ctx context.Context, name string) error

	// GetExtension gets the address of a registered extension. If the extension does
	// not exist, ErrNoExtension is returned.
	GetExtension(ctx context.Context, name string) (*types.Extension, error)

	// GetExtensions gets all the extensions for the namespace in ctx.
	GetExtensions(ctx context.Context, pred *SelectionPredicate) ([]*types.Extension, error)
}
