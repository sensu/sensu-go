package mockstore

import "github.com/sensu/sensu-go/types"

//// Events

// GetEvents ...
func (s *MockStore) GetEvents(org string) ([]*types.Event, error) {
	args := s.Called(org)
	return args.Get(0).([]*types.Event), args.Error(1)
}

// GetEventsByEntity ...
func (s *MockStore) GetEventsByEntity(org, entityID string) ([]*types.Event, error) {
	args := s.Called(org, entityID)
	return args.Get(0).([]*types.Event), args.Error(1)
}

// GetEventByEntityCheck ...
func (s *MockStore) GetEventByEntityCheck(org, entityID, checkID string) (*types.Event, error) {
	args := s.Called(org, entityID, checkID)
	return args.Get(0).(*types.Event), args.Error(1)
}

// UpdateEvent ...
func (s *MockStore) UpdateEvent(event *types.Event) error {
	args := s.Called(event)
	return args.Error(0)
}

// DeleteEventByEntityCheck ...
func (s *MockStore) DeleteEventByEntityCheck(org, entityID, checkID string) error {
	args := s.Called(org, entityID, checkID)
	return args.Error(0)
}
