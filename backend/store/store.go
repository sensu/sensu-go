package store

import "github.com/sensu/sensu-go/types"

// A Store is responsible for managing durable state for Sensu backends.
type Store interface {
	GetEntityByID(id string) (*types.Entity, error)
	UpdateEntity(e *types.Entity) error
	DeleteEntity(e *types.Entity) error
}
