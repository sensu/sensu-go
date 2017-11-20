package store

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

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

	// EventStore provides an interface for managing events
	EventStore

	// EventFilterStore provides an interface for managing events filters
	EventFilterStore

	// HandlerStore provides an interface for managing events handlers
	HandlerStore

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
	GetAssets(ctx context.Context) (assets []*types.Asset, err error)

	// GetAssetByName returns an asset using the given name and the organization
	// stored in ctx. The resulting asset is nil if none was found.
	GetAssetByName(ctx context.Context, name string) (asset *types.Asset, err error)

	// UpdateAsset creates or updates a given asset.
	UpdateAsset(ctx context.Context, asset *types.Asset) error
}

// AuthenticationStore provides methods for managing the JWT secret
type AuthenticationStore interface {
	// CreateJWTSecret create the given JWT secret and returns an error if it was
	// unsuccessful or if a secret already exists.
	CreateJWTSecret(secret []byte) error

	// GetJWTSecret returns the JWT secret.
	GetJWTSecret() (secret []byte, err error)

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
	GetCheckConfigs(ctx context.Context) (checks []*types.CheckConfig, err error)

	// GetCheckConfigByName returns a check's configuration using the given name
	// and the organization and environment stored in ctx. The resulting check is
	// nil if none was found.
	GetCheckConfigByName(ctx context.Context, name string) (*types.CheckConfig, error)

	// UpdateCheckConfig creates or updates a given check's configuration.
	UpdateCheckConfig(ctx context.Context, check *types.CheckConfig) error
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
	GetEntities(ctx context.Context) (entities []*types.Entity, err error)

	// GetEntityByID returns an entity using the given id and the organization
	// and environment stored in ctx. The resulting entity is nil if none was
	// found.
	GetEntityByID(ctx context.Context, id string) (entity *types.Entity, err error)

	// UpdateEntity creates or updates a given entity.
	UpdateEntity(ctx context.Context, entity *types.Entity) error
}

// EnvironmentStore provides methods for managing environments
type EnvironmentStore interface {
	// DeleteEnvironment deletes an environment using the given env struct.
	DeleteEnvironment(ctx context.Context, env *types.Environment) error

	// GetEnvironment returns an environment using the given org and env. The
	// result is nil if none was found.
	GetEnvironment(ctx context.Context, org, env string) (result *types.Environment, err error)

	// GetEnvironments returns all environments in the given ctx's organization. A
	// nil slice with no error is returned if none were found.
	GetEnvironments(ctx context.Context, org string) (envs []*types.Environment, err error)

	// UpdateEnvironment creates or updates a given env.
	UpdateEnvironment(ctx context.Context, env *types.Environment) error
}

// EventStore provides methods for managing events
type EventStore interface {
	// DeleteEventByEntityCheck deletes an event using the given entity and check,
	// within the organization and environment stored in ctx.
	DeleteEventByEntityCheck(ctx context.Context, entity, check string) error

	// GetEvents returns all events in the given ctx's organization and
	// environment. A nil slice with no error is returned if none were found.
	GetEvents(ctx context.Context) (events []*types.Event, err error)

	// GetEventsByEntity returns all events for the given entity within the ctx's
	// organization and environment. A nil slice with no error is returned if none
	// were found.
	GetEventsByEntity(ctx context.Context, entity string) (events []*types.Event, err error)

	// GetEventByEntityCheck returns an event using the given entity and check,
	// within the organization and environment stored in ctx. The resulting event
	// is nil if none was found.
	GetEventByEntityCheck(ctx context.Context, entity, check string) (event *types.Event, err error)

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
	GetEventFilters(ctx context.Context) (filters []*types.EventFilter, err error)

	// GetEventFilterByName returns a filter using the given name and the
	// organization and environment stored in ctx. The resulting filter is nil if
	// none was found.
	GetEventFilterByName(ctx context.Context, name string) (filter *types.EventFilter, err error)

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
	GetHandlers(ctx context.Context) (handlers []*types.Handler, err error)

	// GetHandlerByName returns a handler using the given name and the
	// organization and environment stored in ctx. The resulting handler is nil if
	// none was found.
	GetHandlerByName(ctx context.Context, name string) (handler *types.Handler, err error)

	// UpdateHandler creates or updates a given handler.
	UpdateHandler(ctx context.Context, handler *types.Handler) error
}

// KeepaliveStore provides methods for managing entities keepalives
type KeepaliveStore interface {
	// DeleteFailingKeepalive deletes a failing keepalive record for a given entity.
	DeleteFailingKeepalive(ctx context.Context, entity *types.Entity) error

	// GetFailingKeepalives returns a slice of failing keepalives.
	GetFailingKeepalives(ctx context.Context) (records []*types.KeepaliveRecord, err error)

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
	GetMutators(ctx context.Context) (mutators []*types.Mutator, err error)

	// GetMutatorByName returns a mutator using the given name and the
	// organization and environment stored in ctx. The resulting mutator is nil if
	// none was found.
	GetMutatorByName(ctx context.Context, name string) (mutator *types.Mutator, err error)

	// UpdateMutator creates or updates a given mutator.
	UpdateMutator(ctx context.Context, mutator *types.Mutator) error
}

// OrganizationStore provides methods for managing organizations
type OrganizationStore interface {
	// DeleteOrganizationByName deletes an organization using the given name.
	DeleteOrganizationByName(ctx context.Context, name string) error

	// GetOrganizations returns all organizations. A nil slice with no error is
	// returned if none were found.
	GetOrganizations(ctx context.Context) (orgs []*types.Organization, err error)

	// GetOrganizationByName returns an organization using the given name. The
	// result is nil if none was found.
	GetOrganizationByName(ctx context.Context, name string) (org *types.Organization, err error)

	// UpdateOrganization creates or updates a given organization.
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
	// DeleteSilencedEntriesByCheckName deletes an entry using the given check.
	DeleteSilencedEntriesByCheckName(ctx context.Context, check string) error

	// DeleteSilencedEntriesBySubscription deletes an entry using the given id.
	DeleteSilencedEntriesBySubscription(ctx context.Context, subscription string) error

	// DeleteSilencedEntryByID deletes an entry using the given id.
	DeleteSilencedEntryByID(ctx context.Context, id string) error

	// GetSilencedEntries returns all entries. A nil slice with no error is
	// returned if none were found.
	GetSilencedEntries(ctx context.Context) (entries []*types.Silenced, err error)

	// GetSilencedEntriesByCheckName returns all entries for the given check
	// within the ctx's organization and environment. A nil slice with no error is
	// returned if none were found.
	GetSilencedEntriesByCheckName(ctx context.Context, check string) (entries []*types.Silenced, err error)

	// GetSilencedEntriesByCheckName returns all entries for the given subscription
	// within the ctx's organization and environment. A nil slice with no error is
	// returned if none were found.
	GetSilencedEntriesBySubscription(ctx context.Context, subscription string) (entries []*types.Silenced, err error)

	// GetSilencedEntryByID returns an entry using the given id and the
	// organization and environment stored in ctx. The resulting entry is nil if
	// none was found.
	GetSilencedEntryByID(ctx context.Context, id string) (entry *types.Silenced, err error)

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
	GetToken(subject, id string) (claims *types.Claims, err error)
}

// UserStore provides methods for managing users
type UserStore interface {
	// AuthenticateUser attempts to authenticate a user with the given username
	// and hashed password. An error is returned if the user does not exist, is
	// disabled or the given password does not match.
	AuthenticateUser(ctx context.Context, username, password string) (user *types.User, err error)

	// CreateUsern creates a new user with the given user struct.
	CreateUser(user *types.User) error

	// DeleteUser deletes a user using the given user struct.
	DeleteUser(ctx context.Context, user *types.User) error

	// GetUser returns a user using the given username.
	GetUser(ctx context.Context, username string) (user *types.User, err error)

	// GetUsers returns all enabled users. A nil slice with no error is
	// returned if none were found.
	GetUsers() (users []*types.User, err error)

	// GetUsers returns all users, including the disabled ones. A nil slice with
	// no error is  returned if none were found.
	GetAllUsers() (users []*types.User, err error)

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
