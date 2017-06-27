package store

import (
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

	// Mutators
	MutatorStore

	// Organizations
	OrganizationStore

	// RBAC
	RBACStore

	// Users
	UserStore

	// Keepalives
	KeepaliveStore

	// Initialization of store
	NewInitializer() (Initializer, error)
}

// AssetStore manage assets
type AssetStore interface {
	GetAssets(org string) ([]*types.Asset, error)
	GetAssetByName(org, name string) (*types.Asset, error)
	UpdateAsset(asset *types.Asset) error
	DeleteAssetByName(org, name string) error
}

// AuthenticationStore is responsible for managing the authentication state
type AuthenticationStore interface {
	CreateJWTSecret([]byte) error
	GetJWTSecret() ([]byte, error)
	UpdateJWTSecret([]byte) error
}

// CheckConfigStore provides an interface for interacting & persisting checks
type CheckConfigStore interface {
	GetCheckConfigs(org string) ([]*types.CheckConfig, error)
	GetCheckConfigByName(org, name string) (*types.CheckConfig, error)
	DeleteCheckConfigByName(org, name string) error
	UpdateCheckConfig(check *types.CheckConfig) error
}

// EntityStore provides an interface for interacting & persisting entities
type EntityStore interface {
	GetEntityByID(org, id string) (*types.Entity, error)
	UpdateEntity(e *types.Entity) error
	DeleteEntity(e *types.Entity) error
	DeleteEntityByID(org, id string) error
	GetEntities(org string) ([]*types.Entity, error)
}

// EventStore provides an interface for interacting & persisting events
type EventStore interface {
	GetEvents(org string) ([]*types.Event, error)
	GetEventsByEntity(org, entityID string) ([]*types.Event, error)
	GetEventByEntityCheck(org, entityID, checkID string) (*types.Event, error)
	UpdateEvent(event *types.Event) error
	DeleteEventByEntityCheck(org, entityID, checkID string) error
}

// HandlerStore provides an interface for interacting & persisting handlers
type HandlerStore interface {
	GetHandlers(org string) ([]*types.Handler, error)
	GetHandlerByName(org, name string) (*types.Handler, error)
	DeleteHandlerByName(org, name string) error
	UpdateHandler(handler *types.Handler) error
}

// KeepaliveStore is responsible for updating entity keepalive data.
type KeepaliveStore interface {
	// UpdateKeepalive updates the current expiration time for an entity's
	// keepalive.
	UpdateKeepalive(org, entityID string, expiration int64) error
	// GetKeepalive gets the current expiration for an entity's keepalive.
	GetKeepalive(org, entityID string) (int64, error)
}

// MutatorStore provides an interface for interacting & persisting mutators
type MutatorStore interface {
	GetMutators(org string) ([]*types.Mutator, error)
	GetMutatorByName(org, name string) (*types.Mutator, error)
	DeleteMutatorByName(org, name string) error
	UpdateMutator(mutator *types.Mutator) error
}

// OrganizationStore provides an interface for interacting & persisting orgs
type OrganizationStore interface {
	DeleteOrganizationByName(name string) error
	GetOrganizations() ([]*types.Organization, error)
	GetOrganizationByName(name string) (*types.Organization, error)
	UpdateOrganization(*types.Organization) error
}

// RBACStore provides an interface for interacting & persisting users
type RBACStore interface {
	GetRoles() ([]*types.Role, error)
	GetRoleByName(name string) (*types.Role, error)
	UpdateRole(role *types.Role) error
	DeleteRoleByName(name string) error
}

// UserStore provides an interface for interacting & persisting users
type UserStore interface {
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
