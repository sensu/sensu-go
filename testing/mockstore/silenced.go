package mockstore

import (
	"context"

	corev2 "github.com/sensu/core/v2"
)

// DeleteSilencedEntryByName ...
func (s *MockStore) DeleteSilencedEntryByName(ctx context.Context, silencedID ...string) error {
	args := s.Called(ctx, silencedID)
	return args.Error(0)
}

// GetSilencedEntries ...
func (s *MockStore) GetSilencedEntries(ctx context.Context) ([]*corev2.Silenced, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*corev2.Silenced), args.Error(1)
}

// GetSilencedEntryByName ...
func (s *MockStore) GetSilencedEntryByName(ctx context.Context, silencedID string) (*corev2.Silenced, error) {
	args := s.Called(ctx, silencedID)
	return args.Get(0).(*corev2.Silenced), args.Error(1)
}

func (s *MockStore) GetSilencedEntriesByName(ctx context.Context, names ...string) ([]*corev2.Silenced, error) {
	args := s.Called(ctx, names)
	return args.Get(0).([]*corev2.Silenced), args.Error(1)
}

// GetSilencedEntriesBySubscription ...
func (s *MockStore) GetSilencedEntriesBySubscription(ctx context.Context, subscriptions ...string) ([]*corev2.Silenced, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*corev2.Silenced), args.Error(1)
}

// GetSilencedEntriesByCheckName ...
func (s *MockStore) GetSilencedEntriesByCheckName(ctx context.Context, checkName string) ([]*corev2.Silenced, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*corev2.Silenced), args.Error(1)
}

// UpdateSilencedEntry ...
func (s *MockStore) UpdateSilencedEntry(ctx context.Context, silenced *corev2.Silenced) error {
	args := s.Called(ctx, silenced)
	return args.Error(0)
}
