package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
)

// StoreInitializer ...
type StoreInitializer struct {
	Initialized bool
	Err         error
}

// NewInitializer ...
func (store *MockStore) NewInitializer(context.Context) (store.Initializer, error) {
	return &StoreInitializer{Initialized: false, Err: nil}, nil
}

// Lock ...
func (s *StoreInitializer) Lock(context.Context) error {
	return s.Err
}

// Close ...
func (s *StoreInitializer) Close(context.Context) error {
	return s.Err
}

// IsInitialized ...
func (s *StoreInitializer) IsInitialized(context.Context) (bool, error) {
	return s.Initialized, s.Err
}

// FlagAsInitialized ...
func (s *StoreInitializer) FlagAsInitialized(context.Context) error {
	s.Initialized = true
	return s.Err
}
