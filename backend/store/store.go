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

	// Events
	EventStore

	// Handlers
	HandlerStore

	// Keepalives
	KeepaliveStore

	// Mutators
	MutatorStore

	// Organizations
	OrganizationStore

	// Roles
	RoleStore

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

// EventStore provides an interface for interacting & persisting events
type EventStore interface {
	DeleteEventByEntityCheck(context.Context, string, string) error
	GetEvents(context.Context) ([]*types.Event, error)
	GetEventsByEntity(context.Context, string) ([]*types.Event, error)
	GetEventByEntityCheck(context.Context, string, string) (*types.Event, error)
	UpdateEvent(context.Context, *types.Event) error
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
	// GetKeepalive gets the current expiration for an entity's keepalive.
	GetKeepalive(context.Context, string) (int64, error)
	// UpdateKeepalive updates the current expiration time for an entity's
	// keepalive.
	UpdateKeepalive(context.Context, string, int64) error
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

// RoleStore provides an interface for interacting & persisting users
type RoleStore interface {
	DeleteRoleByName(name string) error
	GetRoles() ([]*types.Role, error)
	GetRoleByName(name string) (*types.Role, error)
	UpdateRole(role *types.Role) error
}

// TokenStore provides an interface for interacting with the JWT whitelist
type TokenStore interface {
	CreateToken(*types.Claims) error
	DeleteToken(string, string) error
	DeleteTokensByUsername(string) error
	GetToken(string, string) (*types.Claims, error)
}

// UserStore provides an interface for interacting & persisting users
type UserStore interface {
	AuthenticateUser(username, password string) (*types.User, error)
	CreateUser(user *types.User) error
	DeleteUserByName(username string) error
	GetUser(username string) (*types.User, error)
	GetUsers() ([]*types.User, error)
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
