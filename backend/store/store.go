package store

import "github.com/sensu/sensu-go/types"

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
}
