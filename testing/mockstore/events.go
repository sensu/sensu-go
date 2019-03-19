package mockstore

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// DeleteEventByEntityCheck ...
func (s *MockStore) DeleteEventByEntityCheck(ctx context.Context, entityName, checkID string) error {
	args := s.Called(ctx, entityName, checkID)
	return args.Error(0)
}

// GetEvents ...
func (s *MockStore) GetEvents(ctx context.Context, pageSize int64, continueToken string) ([]*corev2.Event, string, error) {
	args := s.Called(ctx, pageSize, continueToken)
	return args.Get(0).([]*corev2.Event), args.String(1), args.Error(2)
}

// GetEventsByEntity ...
func (s *MockStore) GetEventsByEntity(ctx context.Context, entityName string) ([]*corev2.Event, error) {
	args := s.Called(ctx, entityName)
	return args.Get(0).([]*corev2.Event), args.Error(1)
}

// GetEventByEntityCheck ...
func (s *MockStore) GetEventByEntityCheck(ctx context.Context, entityName, checkID string) (*corev2.Event, error) {
	args := s.Called(ctx, entityName, checkID)
	return args.Get(0).(*corev2.Event), args.Error(1)
}

// UpdateEvent ...
func (s *MockStore) UpdateEvent(ctx context.Context, event *corev2.Event) error {
	args := s.Called(event)
	return args.Error(0)
}
