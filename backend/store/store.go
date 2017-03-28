package store

import "github.com/sensu/sensu-go/types"

// A Store is responsible for managing durable state for Sensu backends.
type Store interface {
	// Entities
	GetEntityByID(id string) (*types.Entity, error)
	UpdateEntity(e *types.Entity) error
	DeleteEntity(e *types.Entity) error
	GetEntities() ([]*types.Entity, error)

	// Checks
	GetChecks() ([]*types.Check, error)
	GetCheckByName(name string) (*types.Check, error)
	DeleteCheckByName(name string) error
	UpdateCheck(check *types.Check) error
}
