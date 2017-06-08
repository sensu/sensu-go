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

	// Users
	UserStore

	// Keepalives
	KeepaliveStore
}

// AssetStore manage assets
type AssetStore interface {
	GetAssets() ([]*types.Asset, error)
	GetAssetByName(assetName string) (*types.Asset, error)
	UpdateAsset(asset *types.Asset) error
	DeleteAssetByName(assetName string) error
}

// AuthenticationStore is responsible for managing the authentication state
type AuthenticationStore interface {
	CreateJWTSecret([]byte) error
	GetJWTSecret() ([]byte, error)
	UpdateJWTSecret([]byte) error
}

// CheckConfigStore provides an interface for interacting & persisting checks
type CheckConfigStore interface {
	GetCheckConfigs() ([]*types.CheckConfig, error)
	GetCheckConfigByName(name string) (*types.CheckConfig, error)
	DeleteCheckConfigByName(name string) error
	UpdateCheckConfig(check *types.CheckConfig) error
}

// EntityStore provides an interface for interacting & persisting entities
type EntityStore interface {
	GetEntityByID(id string) (*types.Entity, error)
	UpdateEntity(e *types.Entity) error
	DeleteEntity(e *types.Entity) error
	DeleteEntityByID(id string) error
	GetEntities() ([]*types.Entity, error)
}

// EventStore provides an interface for interacting & persisting events
type EventStore interface {
	GetEvents() ([]*types.Event, error)
	GetEventsByEntity(entityID string) ([]*types.Event, error)
	GetEventByEntityCheck(entityID, checkID string) (*types.Event, error)
	UpdateEvent(event *types.Event) error
	DeleteEventByEntityCheck(entityID, checkID string) error
}

// HandlerStore provides an interface for interacting & persisting handlers
type HandlerStore interface {
	GetHandlers() ([]*types.Handler, error)
	GetHandlerByName(name string) (*types.Handler, error)
	DeleteHandlerByName(name string) error
	UpdateHandler(handler *types.Handler) error
}

// MutatorStore provides an interface for interacting & persisting mutators
type MutatorStore interface {
	GetMutators() ([]*types.Mutator, error)
	GetMutatorByName(name string) (*types.Mutator, error)
	DeleteMutatorByName(name string) error
	UpdateMutator(mutator *types.Mutator) error
}

// UserStore provides an interface for interacting & persisting users
type UserStore interface {
	CreateUser(user *types.User) error
	DeleteUserByName(username string) error
	GetUser(username string) (*types.User, error)
	GetUsers() ([]*types.User, error)
	UpdateUser(user *types.User) error
}

// KeepaliveStore is responsible for updating entity keepalive data.
type KeepaliveStore interface {
	// UpdateKeepalive updates the current expiration time for an entity's
	// keepalive.
	UpdateKeepalive(entityID string, expiration int64) error

	// GetKeepalive gets the current expiration for an entity's keepalive.
	GetKeepalive(entityID string) (int64, error)
}
