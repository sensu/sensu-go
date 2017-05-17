package mockstore

import (
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

// MockStore is a store used for testing. When using the MockStore in unit
// tests, stub out the behavior you wish to test against by assigning the
// appropriate function to the appropriate Func field. If you have forgotten
// to stub a particular function, the program will panic.
type MockStore struct {
	mock.Mock
}

// Entities

// GetEntityByID ...
func (s *MockStore) GetEntityByID(id string) (*types.Entity, error) {
	args := s.Called(id)
	return args.Get(0).(*types.Entity), args.Error(1)
}

// UpdateEntity ...
func (s *MockStore) UpdateEntity(e *types.Entity) error {
	args := s.Called(e)
	return args.Error(0)
}

// DeleteEntity ...
func (s *MockStore) DeleteEntity(e *types.Entity) error {
	args := s.Called(e)
	return args.Error(0)
}

// GetEntities ...
func (s *MockStore) GetEntities() ([]*types.Entity, error) {
	args := s.Called()
	return args.Get(0).([]*types.Entity), args.Error(1)
}

// Handlers

// GetHandlers ...
func (s *MockStore) GetHandlers() ([]*types.Handler, error) {
	args := s.Called()
	return args.Get(0).([]*types.Handler), args.Error(1)
}

// GetHandlerByName ...
func (s *MockStore) GetHandlerByName(name string) (*types.Handler, error) {
	args := s.Called(name)
	return args.Get(0).(*types.Handler), args.Error(1)
}

// DeleteHandlerByName ...
func (s *MockStore) DeleteHandlerByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// UpdateHandler ...
func (s *MockStore) UpdateHandler(handler *types.Handler) error {
	args := s.Called(handler)
	return args.Error(0)
}

// Mutators

// GetMutators ...
func (s *MockStore) GetMutators() ([]*types.Mutator, error) {
	args := s.Called()
	return args.Get(0).([]*types.Mutator), args.Error(1)
}

// GetMutatorByName ...
func (s *MockStore) GetMutatorByName(name string) (*types.Mutator, error) {
	args := s.Called(name)
	return args.Get(0).(*types.Mutator), args.Error(1)
}

// DeleteMutatorByName ...
func (s *MockStore) DeleteMutatorByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// UpdateMutator ...
func (s *MockStore) UpdateMutator(mutator *types.Mutator) error {
	args := s.Called(mutator)
	return args.Error(0)
}

// Checks

// GetChecks ...
func (s *MockStore) GetChecks() ([]*types.Check, error) {
	args := s.Called()
	return args.Get(0).([]*types.Check), args.Error(1)
}

// GetCheckByName ...
func (s *MockStore) GetCheckByName(name string) (*types.Check, error) {
	args := s.Called(name)
	return args.Get(0).(*types.Check), args.Error(1)
}

// DeleteCheckByName ...
func (s *MockStore) DeleteCheckByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// UpdateCheck ...
func (s *MockStore) UpdateCheck(check *types.Check) error {
	args := s.Called(check)
	return args.Error(0)
}

// Events

// GetEvents ...
func (s *MockStore) GetEvents() ([]*types.Event, error) {
	args := s.Called()
	return args.Get(0).([]*types.Event), args.Error(1)
}

// GetEventsByEntity ...
func (s *MockStore) GetEventsByEntity(entityID string) ([]*types.Event, error) {
	args := s.Called(entityID)
	return args.Get(0).([]*types.Event), args.Error(1)
}

// GetEventByEntityCheck ...
func (s *MockStore) GetEventByEntityCheck(entityID, checkID string) (*types.Event, error) {
	args := s.Called(entityID, checkID)
	return args.Get(0).(*types.Event), args.Error(1)
}

// UpdateEvent ...
func (s *MockStore) UpdateEvent(event *types.Event) error {
	args := s.Called(event)
	return args.Error(0)
}

// DeleteEventByEntityCheck ...
func (s *MockStore) DeleteEventByEntityCheck(entityID, checkID string) error {
	args := s.Called(entityID, checkID)
	return args.Error(0)
}

// UpdateKeepalive ...
func (s *MockStore) UpdateKeepalive(entityID string, expiration int64) error {
	args := s.Called(entityID, expiration)
	return args.Error(0)
}

// GetKeepalive ...
func (s *MockStore) GetKeepalive(entityID string) (int64, error) {
	args := s.Called(entityID)
	return args.Get(0).(int64), args.Error(1)
}

// UpdateUser ...
func (s *MockStore) UpdateUser(user *types.User) error {
	args := s.Called(user)
	return args.Error(0)
}
