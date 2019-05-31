package mockstore

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteEventByEntityCheck ...
func (s *MockStore) DeleteEventByEntityCheck(ctx context.Context, entityName, checkID string) error {
	args := s.Called(ctx, entityName, checkID)
	return args.Error(0)
}

// GetEvents ...
func (s *MockStore) GetEvents(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*corev2.Event), args.Error(1)
}

// GetEventsByEntity ...
func (s *MockStore) GetEventsByEntity(ctx context.Context, entityName string, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	args := s.Called(ctx, entityName, pred)
	return args.Get(0).([]*corev2.Event), args.Error(1)
}

// GetEventByEntityCheck ...
func (s *MockStore) GetEventByEntityCheck(ctx context.Context, entityName, checkID string) (*corev2.Event, error) {
	args := s.Called(ctx, entityName, checkID)
	return args.Get(0).(*corev2.Event), args.Error(1)
}

// UpdateEvent ...
func (s *MockStore) UpdateEvent(ctx context.Context, event *corev2.Event) (*corev2.Event, *corev2.Event, error) {
	args := s.Called(event)
	return args.Get(0).(*corev2.Event), args.Get(1).(*corev2.Event), args.Error(2)
}
