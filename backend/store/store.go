package store

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// A Store is responsible for managing durable state for Sensu backends.
type Store interface {
	// Assets
	AssetStore

	// Authentication
	AuthenticationStore

	// CheckConfigurations
	CheckConfigStore

	// Entities
	EntityStore

	// Environments
	EnvironmentStore

	// Events
	EventStore

	// Event Filters
	EventFilterStore

	// Handlers
	HandlerStore

	// Keepalives
	KeepaliveStore

	// Mutators
	MutatorStore

	// Organizations
	OrganizationStore

	// RBAC
	RBACStore

	// Silenced Entries
	SilencedStore

	// Tokens
	TokenStore

	// Users
	UserStore

	// Initialization of store
	NewInitializer() (Initializer, error)
}

// AssetStore manage assets
type AssetStore interface {
	DeleteAssetByName(context.Context, string) error
	GetAssets(context.Context) ([]*types.Asset, error)
	GetAssetByName(context.Context, string) (*types.Asset, error)
	UpdateAsset(context.Context, *types.Asset) error
}

// AuthenticationStore is responsible for managing the authentication state
type AuthenticationStore interface {
	CreateJWTSecret([]byte) error
	GetJWTSecret() ([]byte, error)
	UpdateJWTSecret([]byte) error
}

// CheckConfigStore provides an interface for interacting & persisting checks
type CheckConfigStore interface {
	DeleteCheckConfigByName(context.Context, string) error
	GetCheckConfigs(context.Context) ([]*types.CheckConfig, error)
	GetCheckConfigByName(context.Context, string) (*types.CheckConfig, error)
	UpdateCheckConfig(context.Context, *types.CheckConfig) error
}

// EntityStore provides an interface for interacting & persisting entities
type EntityStore interface {
	DeleteEntity(context.Context, *types.Entity) error
	DeleteEntityByID(context.Context, string) error
	GetEntities(context.Context) ([]*types.Entity, error)
	GetEntityByID(context.Context, string) (*types.Entity, error)
	UpdateEntity(context.Context, *types.Entity) error
}

// EnvironmentStore provides an interface for interacting & persisting environments
type EnvironmentStore interface {
	DeleteEnvironment(context.Context, string, string) error
	GetEnvironment(context.Context, string, string) (*types.Environment, error)
	GetEnvironments(context.Context, string) ([]*types.Environment, error)
	UpdateEnvironment(context.Context, string, *types.Environment) error
}

// EventStore provides an interface for interacting & persisting events
type EventStore interface {
	DeleteEventByEntityCheck(context.Context, string, string) error
	GetEvents(context.Context) ([]*types.Event, error)
	GetEventsByEntity(context.Context, string) ([]*types.Event, error)
	GetEventByEntityCheck(context.Context, string, string) (*types.Event, error)
	UpdateEvent(context.Context, *types.Event) error
}

// EventFilterStore provides an interface for interacting & persisting event filters
type EventFilterStore interface {
	DeleteEventFilterByName(context.Context, string) error
	GetEventFilters(context.Context) ([]*types.EventFilter, error)
	GetEventFilterByName(context.Context, string) (*types.EventFilter, error)
	UpdateEventFilter(context.Context, *types.EventFilter) error
}

// HandlerStore provides an interface for interacting & persisting handlers
type HandlerStore interface {
	DeleteHandlerByName(context.Context, string) error
	GetHandlers(context.Context) ([]*types.Handler, error)
	GetHandlerByName(context.Context, string) (*types.Handler, error)
	UpdateHandler(context.Context, *types.Handler) error
}

// KeepaliveStore is responsible for updating entity keepalive data.
type KeepaliveStore interface {
	// DeleteFailingKeepalive deletes a failing keepalive record for an entity.
	DeleteFailingKeepalive(ctx context.Context, entity *types.Entity) error
	// GetFailingKeepalives gets the list of failing keepalives for a given
	// backend.
	GetFailingKeepalives(context.Context) ([]*types.KeepaliveRecord, error)
	// UpdateKeepalive updates the current expiration time for an entity's
	// keepalive.
	UpdateFailingKeepalive(context.Context, *types.Entity, int64) error
}

// MutatorStore provides an interface for interacting & persisting mutators
type MutatorStore interface {
	DeleteMutatorByName(context.Context, string) error
	GetMutators(context.Context) ([]*types.Mutator, error)
	GetMutatorByName(context.Context, string) (*types.Mutator, error)
	UpdateMutator(context.Context, *types.Mutator) error
}

// OrganizationStore provides an interface for interacting & persisting orgs
type OrganizationStore interface {
	DeleteOrganizationByName(context.Context, string) error
	GetOrganizations(context.Context) ([]*types.Organization, error)
	GetOrganizationByName(context.Context, string) (*types.Organization, error)
	UpdateOrganization(context.Context, *types.Organization) error
}

// RBACStore provides an interface for interacting & persisting users
type RBACStore interface {
	GetRoles() ([]*types.Role, error)
	GetRoleByName(name string) (*types.Role, error)
	UpdateRole(role *types.Role) error
	DeleteRoleByName(name string) error
}

// SilencedStore provides an interface for interacting and persisting silenced
// event entries.
type SilencedStore interface {
	GetSilencedEntries(ctx context.Context) ([]*types.Silenced, error)
	GetSilencedEntriesByCheckName(ctx context.Context, checkName string) ([]*types.Silenced, error)
	GetSilencedEntryByID(ctx context.Context, silencedID string) (*types.Silenced, error)
	GetSilencedEntriesBySubscription(ctx context.Context, subscription string) ([]*types.Silenced, error)
	DeleteSilencedEntryByID(ctx context.Context, silencedID string) error
	DeleteSilencedEntriesBySubscription(ctx context.Context, subscription string) error
	DeleteSilencedEntriesByCheckName(ctx context.Context, checkName string) error
	UpdateSilencedEntry(ctx context.Context, silenced *types.Silenced) error
}

// TokenStore provides an interface for interacting with the JWT access list
type TokenStore interface {
	CreateToken(*types.Claims) error
	DeleteTokens(string, []string) error
	GetToken(string, string) (*types.Claims, error)
}

// UserStore provides an interface for interacting & persisting users
type UserStore interface {
	AuthenticateUser(ctx context.Context, username, password string) (*types.User, error)
	CreateUser(user *types.User) error
	DeleteUser(context.Context, *types.User) error
	GetUser(context.Context, string) (*types.User, error)
	GetUsers() ([]*types.User, error)
	GetAllUsers() ([]*types.User, error)
	UpdateUser(user *types.User) error
}

// Initializer utility provides way to determine if store is initialized
// and mechanism for setting it to the initialized state.
type Initializer interface {
	Lock() error
	Close() error
	IsInitialized() (bool, error)
	FlagAsInitialized() error
}
