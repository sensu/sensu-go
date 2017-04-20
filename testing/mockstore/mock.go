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

	// Entities
	FuncGetEntityByID func(id string) (*types.Entity, error)
	FuncUpdateEntity  func(e *types.Entity) error
	FuncDeleteEntity  func(e *types.Entity) error
	FuncGetEntities   func() ([]*types.Entity, error)

	// Handlers
	FuncGetHandlers         func() ([]*types.Handler, error)
	FuncGetHandlerByName    func(name string) (*types.Handler, error)
	FuncDeleteHandlerByName func(name string) error
	FuncUpdateHandler       func(handler *types.Handler) error

	// Mutators
	FuncGetMutators         func() ([]*types.Mutator, error)
	FuncGetMutatorByName    func(name string) (*types.Mutator, error)
	FuncDeleteMutatorByName func(name string) error
	FuncUpdateMutator       func(mutator *types.Mutator) error

	// Checks
	FuncGetChecks         func() ([]*types.Check, error)
	FuncGetCheckByName    func(name string) (*types.Check, error)
	FuncDeleteCheckByName func(name string) error
	FuncUpdateCheck       func(check *types.Check) error

	// Events
	FuncGetEvents                func() ([]*types.Event, error)
	FuncGetEventsByEntity        func(entityID string) ([]*types.Event, error)
	FuncGetEventByEntityCheck    func(entityID, checkID string) (*types.Event, error)
	FuncUpdateEvent              func(event *types.Event) error
	FuncDeleteEventByEntityCheck func(entityID, checkID string) error
}

// Entities
func (s *MockStore) GetEntityByID(id string) (*types.Entity, error) {
	if s.FuncGetEntityByID != nil {
		return s.FuncGetEntityByID(id)
	}
	args := s.Called(id)
	return args.Get(0).(*types.Entity), args.Error(1)
}

func (s *MockStore) UpdateEntity(e *types.Entity) error {
	if s.FuncUpdateEntity != nil {
		return s.FuncUpdateEntity(e)
	}
	args := s.Called(e)
	return args.Error(0)
}

func (s *MockStore) DeleteEntity(e *types.Entity) error {
	if s.FuncDeleteEntity != nil {
		return s.FuncDeleteEntity(e)
	}
	args := s.Called(e)
	return args.Error(0)
}

func (s *MockStore) GetEntities() ([]*types.Entity, error) {
	if s.FuncGetEntities != nil {
		return s.FuncGetEntities()
	}
	args := s.Called()
	return args.Get(0).([]*types.Entity), args.Error(1)
}

// Handlers
func (s *MockStore) GetHandlers() ([]*types.Handler, error) {
	if s.FuncGetHandlers != nil {
		return s.FuncGetHandlers()
	}
	args := s.Called()
	return args.Get(0).([]*types.Handler), args.Error(1)
}

func (s *MockStore) GetHandlerByName(name string) (*types.Handler, error) {
	if s.FuncGetHandlerByName != nil {
		return s.FuncGetHandlerByName(name)
	}
	args := s.Called(name)
	return args.Get(0).(*types.Handler), args.Error(1)
}

func (s *MockStore) DeleteHandlerByName(name string) error {
	if s.FuncDeleteHandlerByName != nil {
		return s.FuncDeleteHandlerByName(name)
	}
	args := s.Called(name)
	return args.Error(0)
}

func (s *MockStore) UpdateHandler(handler *types.Handler) error {
	if s.FuncUpdateHandler != nil {
		return s.FuncUpdateHandler(handler)
	}
	args := s.Called(handler)
	return args.Error(0)
}

// Mutators
func (s *MockStore) GetMutators() ([]*types.Mutator, error) {
	if s.FuncGetMutators != nil {
		return s.FuncGetMutators()
	}
	args := s.Called()
	return args.Get(0).([]*types.Mutator), args.Error(1)
}

func (s *MockStore) GetMutatorByName(name string) (*types.Mutator, error) {
	if s.FuncGetMutatorByName != nil {
		return s.FuncGetMutatorByName(name)
	}
	args := s.Called(name)
	return args.Get(0).(*types.Mutator), args.Error(1)
}

func (s *MockStore) DeleteMutatorByName(name string) error {
	if s.FuncDeleteMutatorByName != nil {
		return s.FuncDeleteMutatorByName(name)
	}
	args := s.Called(name)
	return args.Error(0)
}

func (s *MockStore) UpdateMutator(mutator *types.Mutator) error {
	if s.FuncUpdateMutator != nil {
		return s.FuncUpdateMutator(mutator)
	}
	args := s.Called(mutator)
	return args.Error(0)
}

// Checks
func (s *MockStore) GetChecks() ([]*types.Check, error) {
	if s.FuncGetChecks != nil {
		return s.FuncGetChecks()
	}
	args := s.Called()
	return args.Get(0).([]*types.Check), args.Error(1)
}

func (s *MockStore) GetCheckByName(name string) (*types.Check, error) {
	if s.FuncGetCheckByName != nil {
		return s.FuncGetCheckByName(name)
	}
	args := s.Called(name)
	return args.Get(0).(*types.Check), args.Error(1)
}

func (s *MockStore) DeleteCheckByName(name string) error {
	if s.FuncDeleteCheckByName != nil {
		return s.FuncDeleteCheckByName(name)
	}
	args := s.Called(name)
	return args.Error(0)
}

func (s *MockStore) UpdateCheck(check *types.Check) error {
	if s.FuncUpdateCheck != nil {
		return s.UpdateCheck(check)
	}
	args := s.Called(check)
	return args.Error(0)
}

// Events
func (s *MockStore) GetEvents() ([]*types.Event, error) {
	if s.FuncGetEvents != nil {
		return s.FuncGetEvents()
	}
	args := s.Called()
	return args.Get(0).([]*types.Event), args.Error(1)
}

func (s *MockStore) GetEventsByEntity(entityID string) ([]*types.Event, error) {
	if s.FuncGetEventsByEntity != nil {
		return s.FuncGetEventsByEntity(entityID)
	}
	args := s.Called(entityID)
	return args.Get(0).([]*types.Event), args.Error(1)
}

func (s *MockStore) GetEventByEntityCheck(entityID, checkID string) (*types.Event, error) {
	if s.FuncGetEventByEntityCheck != nil {
		return s.FuncGetEventByEntityCheck(entityID, checkID)
	}
	args := s.Called(entityID, checkID)
	return args.Get(0).(*types.Event), args.Error(1)
}

func (s *MockStore) UpdateEvent(event *types.Event) error {
	if s.FuncUpdateEvent != nil {
		return s.FuncUpdateEvent(event)
	}
	args := s.Called(event)
	return args.Error(0)
}

func (s *MockStore) DeleteEventByEntityCheck(entityID, checkID string) error {
	if s.FuncDeleteEventByEntityCheck != nil {
		return s.FuncDeleteEventByEntityCheck(entityID, checkID)
	}
	args := s.Called(entityID, checkID)
	return args.Error(0)
}
