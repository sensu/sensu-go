package mockstore

import "github.com/sensu/sensu-go/backend/store"

// StoreInitializer ...
type StoreInitializer struct {
	Initialized bool
	Err         error
}

// NewInitializer ...
func (store *MockStore) NewInitializer() (store.Initializer, error) {
	return &StoreInitializer{Initialized: false, Err: nil}, nil
}

// Lock ...
func (s *StoreInitializer) Lock() error {
	return s.Err
}

// Close ...
func (s *StoreInitializer) Close() error {
	return s.Err
}

// IsInitialized ...
func (s *StoreInitializer) IsInitialized() (bool, error) {
	return s.Initialized, s.Err
}

// FlagAsInitialized ...
func (s *StoreInitializer) FlagAsInitialized() error {
	s.Initialized = true
	return s.Err
}
