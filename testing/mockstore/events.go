package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// DeleteEventByEntityCheck ...
func (s *MockStore) DeleteEventByEntityCheck(ctx context.Context, entityName, checkID string) error {
	args := s.Called(ctx, entityName, checkID)
	return args.Error(0)
}

// GetEvents ...
func (s *MockStore) GetEvents(ctx context.Context) ([]*types.Event, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*types.Event), args.Error(1)
}

// GetEventsByEntity ...
func (s *MockStore) GetEventsByEntity(ctx context.Context, entityName string) ([]*types.Event, error) {
	args := s.Called(ctx, entityName)
	return args.Get(0).([]*types.Event), args.Error(1)
}

// GetEventByEntityCheck ...
func (s *MockStore) GetEventByEntityCheck(ctx context.Context, entityName, checkID string) (*types.Event, error) {
	args := s.Called(ctx, entityName, checkID)
	return args.Get(0).(*types.Event), args.Error(1)
}

// UpdateEvent ...
func (s *MockStore) UpdateEvent(ctx context.Context, event *types.Event) error {
	args := s.Called(event)
	return args.Error(0)
}
