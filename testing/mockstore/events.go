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
func (s *MockStore) GetEvents(ctx context.Context, pageSize int64, continueToken string) ([]*types.Event, string, error) {
	args := s.Called(ctx, pageSize, continueToken)
	return args.Get(0).([]*types.Event), args.String(1), args.Error(2)
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
