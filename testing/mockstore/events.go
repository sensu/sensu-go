package mockstore

import "github.com/sensu/sensu-go/types"

//// Events

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
