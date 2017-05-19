package store

import (
	"github.com/sensu/sensu-go/types"
)

// A Store is responsible for managing durable state for Sensu backends.
type Store interface {
	// Entities
	GetEntityByID(id string) (*types.Entity, error)
	UpdateEntity(e *types.Entity) error
	DeleteEntity(e *types.Entity) error
	GetEntities() ([]*types.Entity, error)

	// Handlers
	GetHandlers() ([]*types.Handler, error)
	GetHandlerByName(name string) (*types.Handler, error)
	DeleteHandlerByName(name string) error
	UpdateHandler(handler *types.Handler) error

	// Mutators
	GetMutators() ([]*types.Mutator, error)
	GetMutatorByName(name string) (*types.Mutator, error)
	DeleteMutatorByName(name string) error
	UpdateMutator(mutator *types.Mutator) error

	// Checks
	GetChecks() ([]*types.Check, error)
	GetCheckByName(name string) (*types.Check, error)
	DeleteCheckByName(name string) error
	UpdateCheck(check *types.Check) error

	// Events
	GetEvents() ([]*types.Event, error)
	GetEventsByEntity(entityID string) ([]*types.Event, error)
	GetEventByEntityCheck(entityID, checkID string) (*types.Event, error)
	UpdateEvent(event *types.Event) error
	DeleteEventByEntityCheck(entityID, checkID string) error

	// Users
	CreateUser(user *types.User) error
	GetUser(username string) (*types.User, error)
	UpdateUser(user *types.User) error

	// Assets
	AssetStore
}

// AssetStore manage assets
type AssetStore interface {
	GetAssets() ([]*types.Asset, error)
	GetAssetByName(assetName string) (*types.Asset, error)
	UpdateAsset(asset *types.Asset) error
	DeleteAssetByName(assetName string) error

	KeepaliveStore
}

// KeepaliveStore is responsible for updating entity keepalive data.
type KeepaliveStore interface {
	// UpdateKeepalive updates the current expiration time for an entity's
	// keepalive.
	UpdateKeepalive(entityID string, expiration int64) error

	// GetKeepalive gets the current expiration for an entity's keepalive.
	GetKeepalive(entityID string) (int64, error)
}
