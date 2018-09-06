package store

import (
	"context"
	"errors"

	"github.com/sensu/sensu-go/types"
)

const (
	// WatchUnknown indicates that we received an unknown watch even tytpe
	// from etcd.
	WatchUnknown WatchActionType = iota
	// WatchCreate indicates that an object was created.
	WatchCreate
	// WatchUpdate indicates that an object was updated.
	WatchUpdate
	// WatchDelete indicates that an object was deleted.
	WatchDelete
)

// WatchActionType indicates what type of change was made to an object in the store.
type WatchActionType int

func (t WatchActionType) String() string {
	var s string
	switch t {
	case WatchUnknown:
		s = "Unknown"
	case WatchCreate:
		s = "Create"
	case WatchDelete:
		s = "Delete"
	case WatchUpdate:
		s = "Update"
	}
	return s
}

// A WatchEventCheckConfig contains the modified store object and the action that occured
// during the modification.
type WatchEventCheckConfig struct {
	CheckConfig *types.CheckConfig
	Action      WatchActionType
}

// A WatchEventAsset contains the modified asset object and the action that occurred
// during the modification.
type WatchEventAsset struct {
	Asset  *types.Asset
	Action WatchActionType
}

// A WatchEventHookConfig contains the modified asset object and the action that occurred
// during the modification.
type WatchEventHookConfig struct {
	HookConfig *types.HookConfig
	Action     WatchActionType
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

	// EntityStore provides an interface for managing entities
	EntityStore

	// EnvironmentStore provides an interface for managing environments
	EnvironmentStore

	// ErrorStore provides an interface for managing pipeline errors
	ErrorStore

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

	// OrganizationStore provides an interface for managing organizations
	OrganizationStore

	// RBACStore provides an interface for managing RBAC roles and rules
	RBACStore

	// SilencedStore provides an interface for managing silenced entries,
	// consisting of entities, subscriptions and/or checks
	SilencedStore

	// TokenStore provides an interface for managing the JWT access list
	TokenStore

	// UserStore provides an interface for managing users
	UserStore

	// ExtensionRegistry tracks third-party extensions.
	ExtensionRegistry

	// NewInitializer returns the Initializer interfaces, which provides the
	// required mechanism to verify if a store is initialized
	NewInitializer() (Initializer, error)
}

// AssetStore provides methods for managing checks assets
type AssetStore interface {
	// DeleteAssetByName deletes an asset using the given name and the
	// organization stored in ctx.
	DeleteAssetByName(ctx context.Context, name string) error

	// GetAssets returns all assets in the given ctx's organization. A nil
	// slice with no error is returned if none were found.
	GetAssets(ctx context.Context) ([]*types.Asset, error)

	// GetAssetByName returns an asset using the given name and the organization
	// stored in ctx. The resulting asset is nil if none was found.
	GetAssetByName(ctx context.Context, name string) (*types.Asset, error)

	// UpdateAsset creates or updates a given asset.
	UpdateAsset(ctx context.Context, asset *types.Asset) error

	// GetAssetWatcher returns a channel that emits WatchEventAsset structs notifying
	// the caller that an Asset was updated. If the watcher runs into a terminal error
	// or the context passed is cancelled, then the channel will be closed. The caller must
	// restart the watcher, if needed.
	GetAssetWatcher(ctx context.Context) <-chan WatchEventAsset
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
	// and the organization and environment stored in ctx.
	DeleteCheckConfigByName(ctx context.Context, name string) error

	// GetCheckConfigs returns all checks configurations in the given ctx's
	// organization and environment. A nil slice with no error is returned if none
	// were found.
	GetCheckConfigs(ctx context.Context) ([]*types.CheckConfig, error)

	// GetCheckConfigByName returns a check's configuration using the given name
	// and the organization and environment stored in ctx. The resulting check is
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

// HookConfigStore provides methods for managing hooks configuration
type HookConfigStore interface {
	// DeleteHookConfigByName deletes a hook's configuration using the given name
	// and the organization and environment stored in ctx.
	DeleteHookConfigByName(ctx context.Context, name string) error

	// GetHookConfigs returns all hooks configurations in the given ctx's
	// organization and environment. A nil slice with no error is returned if none
	// were found.
	GetHookConfigs(ctx context.Context) ([]*types.HookConfig, error)

	// GetHookConfigByName returns a hook's configuration using the given name
	// and the organization and environment stored in ctx. The resulting hook is
	// nil if none was found.
	GetHookConfigByName(ctx context.Context, name string) (*types.HookConfig, error)

	// UpdateHookConfig creates or updates a given hook's configuration.
	UpdateHookConfig(ctx context.Context, check *types.HookConfig) error

	// GetHookConfigWatcher returns a channel that emits WatchEventHookConfig structs notifying
	// the caller that a HookConfig was updated. If the watcher runs into a terminal error
	// or the context passed is cancelled, then the channel will be closed. The caller must
	// restart the watcher, if needed.
	GetHookConfigWatcher(ctx context.Context) <-chan WatchEventHookConfig
}

// EntityStore provides methods for managing entities
type EntityStore interface {
	// DeleteEntity deletes an entity using the given entity struct.
	DeleteEntity(ctx context.Context, entity *types.Entity) error

	// DeleteEntityByID deletes an entity using the given id and the
	// organization and environment stored in ctx.
	DeleteEntityByID(ctx context.Context, id string) error

	// GetEntities returns all entities in the given ctx's organization and
	// environment. A nil slice with no error is returned if none were found.
	GetEntities(ctx context.Context) ([]*types.Entity, error)

	// GetEntityByID returns an entity using the given id and the organization
	// and environment stored in ctx. The resulting entity is nil if none was
	// found.
	GetEntityByID(ctx context.Context, id string) (*types.Entity, error)

	// UpdateEntity creates or updates a given entity.
	UpdateEntity(ctx context.Context, entity *types.Entity) error
}

// EnvironmentStore provides methods for managing environments
type EnvironmentStore interface {
	// DeleteEnvironment deletes an environment using the given env struct.
	DeleteEnvironment(ctx context.Context, env *types.Environment) error

	// GetEnvironment returns an environment using the given org and env. The
	// result is nil if none was found.
	GetEnvironment(ctx context.Context, org, env string) (*types.Environment, error)

	// GetEnvironments returns all environments in the given ctx's organization. A
	// nil slice with no error is returned if none were found.
	GetEnvironments(ctx context.Context, org string) ([]*types.Environment, error)

	// UpdateEnvironment creates or updates a given env.
	UpdateEnvironment(ctx context.Context, env *types.Environment) error
}

// ErrorStore provides methods for managing pipeline errors
type ErrorStore interface {
	// DeleteError deletes an error using the given entity, check and timestamp,
	// within the organization and environment stored in ctx.
	DeleteError(ctx context.Context, entity, check, timestamp string) error

	// DeleteErrorsByEntity deletes all errors associated with the given entity,
	// within the organization and environment stored in ctx.
	DeleteErrorsByEntity(ctx context.Context, entity string) error

	// DeleteErrorsByEntityCheck deletes all errors associated with the given
	// entity and check within the organization and environment stored in ctx.
	DeleteErrorsByEntityCheck(ctx context.Context, entity, check string) error

	// GetError returns error associated with given entity, check and timestamp,
	// in the given ctx's organization and environment.
	GetError(ctx context.Context, entity, check, timestamp string) (*types.Error, error)

	// GetErrors returns all errors in the given ctx's organization and
	// environment.
	GetErrors(ctx context.Context) ([]*types.Error, error)

	// GetErrorsByEntity returns all errors for the given entity within the ctx's
	// organization and environment. A nil slice with no error is returned if none
	// were found.
	GetErrorsByEntity(ctx context.Context, entity string) ([]*types.Error, error)

	// GetErrorByEntityCheck returns an error using the given entity and check,
	// within the organization and environment stored in ctx. The resulting error
	// is nil if none was found.
	GetErrorsByEntityCheck(ctx context.Context, entity, check string) ([]*types.Error, error)

	// CreateError creates or updates a given error.
	CreateError(ctx context.Context, error *types.Error) error
}

// EventStore provides methods for managing events
type EventStore interface {
	// DeleteEventByEntityCheck deletes an event using the given entity and check,
	// within the organization and environment stored in ctx.
	DeleteEventByEntityCheck(ctx context.Context, entity, check string) error

	// GetEvents returns all events in the given ctx's organization and
	// environment. A nil slice with no error is returned if none were found.
	GetEvents(ctx context.Context) ([]*types.Event, error)

	// GetEventsByEntity returns all events for the given entity within the ctx's
	// organization and environment. A nil slice with no error is returned if none
	// were found.
	GetEventsByEntity(ctx context.Context, entity string) ([]*types.Event, error)

	// GetEventByEntityCheck returns an event using the given entity and check,
	// within the organization and environment stored in ctx. The resulting event
	// is nil if none was found.
	GetEventByEntityCheck(ctx context.Context, entity, check string) (*types.Event, error)

	// UpdateEvent creates or updates a given event.
	UpdateEvent(ctx context.Context, event *types.Event) error
}

// EventFilterStore provides methods for managing events filters
type EventFilterStore interface {
	// DeleteEventFilterByName deletes an event filter using the given name and the
	// organization and environment stored in ctx.
	DeleteEventFilterByName(ctx context.Context, name string) error

	// GetEventFilters returns all filters in the given ctx's organization and
	// environment. A nil slice with no error is returned if none were found.
	GetEventFilters(ctx context.Context) ([]*types.EventFilter, error)

	// GetEventFilterByName returns a filter using the given name and the
	// organization and environment stored in ctx. The resulting filter is nil if
	// none was found.
	GetEventFilterByName(ctx context.Context, name string) (*types.EventFilter, error)

	// UpdateEventFilter creates or updates a given filter.
	UpdateEventFilter(ctx context.Context, filter *types.EventFilter) error
}

// HandlerStore provides methods for managing events handlers
type HandlerStore interface {
	// DeleteHandlerByName deletes a handler using the given name and the
	// organization and environment stored in ctx.
	DeleteHandlerByName(ctx context.Context, name string) error

	// GetHandlers returns all handlers in the given ctx's organization and
	// environment. A nil slice with no error is returned if none were found.
	GetHandlers(ctx context.Context) ([]*types.Handler, error)

	// GetHandlerByName returns a handler using the given name and the
	// organization and environment stored in ctx. The resulting handler is nil if
	// none was found.
	GetHandlerByName(ctx context.Context, name string) (*types.Handler, error)

	// UpdateHandler creates or updates a given handler.
	UpdateHandler(ctx context.Context, handler *types.Handler) error
}

// HealthStore provides methods for cluster health
type HealthStore interface {
	GetClusterHealth(ctx context.Context) *types.HealthResponse
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
	// organization and environment stored in ctx.
	DeleteMutatorByName(ctx context.Context, name string) error

	// GetMutators returns all mutators in the given ctx's organization and
	// environment. A nil slice with no error is returned if none were found.
	GetMutators(ctx context.Context) ([]*types.Mutator, error)

	// GetMutatorByName returns a mutator using the given name and the
	// organization and environment stored in ctx. The resulting mutator is nil if
	// none was found.
	GetMutatorByName(ctx context.Context, name string) (*types.Mutator, error)

	// UpdateMutator creates or updates a given mutator.
	UpdateMutator(ctx context.Context, mutator *types.Mutator) error
}

// OrganizationStore provides methods for managing organizations
type OrganizationStore interface {
	// CreateOrganization creates a given organization and a default environment
	// within this new organization
	CreateOrganization(ctx context.Context, org *types.Organization) error

	// DeleteOrganizationByName deletes an organization using the given name.
	DeleteOrganizationByName(ctx context.Context, name string) error

	// GetOrganizations returns all organizations. A nil slice with no error is
	// returned if none were found.
	GetOrganizations(ctx context.Context) ([]*types.Organization, error)

	// GetOrganizationByName returns an organization using the given name. The
	// result is nil if none was found.
	GetOrganizationByName(ctx context.Context, name string) (*types.Organization, error)

	// UpdateOrganization updates an existing organization.
	UpdateOrganization(ctx context.Context, org *types.Organization) error
}

// RBACStore provides methods for managing RBAC roles and rules
type RBACStore interface {
	// DeleteRoleByName deletes a role using the given name.
	DeleteRoleByName(ctx context.Context, name string) error

	// GetRoleByName returns a role using the given name. The result is nil if
	// none was found.
	GetRoleByName(ctx context.Context, name string) (*types.Role, error)

	// GetRoles returns all roles. A nil slice with no error is returned if none
	// were found.
	GetRoles(context.Context) ([]*types.Role, error)

	// UpdateRole creates or updates a given role.
	UpdateRole(ctx context.Context, role *types.Role) error
}

// SilencedStore provides methods for managing silenced entries,
// consisting of entities, subscriptions and/or checks
type SilencedStore interface {
	// DeleteSilencedEntryByID deletes an entry using the given id.
	DeleteSilencedEntryByID(ctx context.Context, id string) error

	// GetSilencedEntries returns all entries. A nil slice with no error is
	// returned if none were found.
	GetSilencedEntries(ctx context.Context) ([]*types.Silenced, error)

	// GetSilencedEntriesByCheckName returns all entries for the given check
	// within the ctx's organization and environment. A nil slice with no error is
	// returned if none were found.
	GetSilencedEntriesByCheckName(ctx context.Context, check string) ([]*types.Silenced, error)

	// GetSilencedEntriesByCheckName returns all entries for the given subscription
	// within the ctx's organization and environment. A nil slice with no error is
	// returned if none were found.
	GetSilencedEntriesBySubscription(ctx context.Context, subscription string) ([]*types.Silenced, error)

	// GetSilencedEntryByID returns an entry using the given id and the
	// organization and environment stored in ctx. The resulting entry is nil if
	// none was found.
	GetSilencedEntryByID(ctx context.Context, id string) (*types.Silenced, error)

	// UpdateHandler creates or updates a given entry.
	UpdateSilencedEntry(ctx context.Context, entry *types.Silenced) error
}

// TokenStore provides methods for managing the JWT access list
type TokenStore interface {
	// CreateToken creates a new entry in the JWT access list with the given claims.
	CreateToken(claims *types.Claims) error

	// DeleteTokens deletes one or multiple given tokens, belonging to the same
	// given subject.
	DeleteTokens(subject string, tokens []string) error

	// GetToken returns the claims of a given token ID, belonging to the given
	// subject. An error is returned if no claims were found.
	GetToken(subject, id string) (*types.Claims, error)
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
	GetAllUsers() ([]*types.User, error)

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

	// GetExtensions gets all the extensions for the organization in ctx.
	GetExtensions(ctx context.Context) ([]*types.Extension, error)
}
