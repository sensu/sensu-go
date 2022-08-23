package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

func (s *MockStore) DeleteSilences(ctx context.Context, namespace string, silencedID []string) error {
	args := s.Called(ctx, namespace, silencedID)
	return args.Error(0)
}

func (s *MockStore) GetSilences(ctx context.Context, namespace string) ([]*types.Silenced, error) {
	args := s.Called(ctx, namespace)
	return args.Get(0).([]*types.Silenced), args.Error(1)
}

func (s *MockStore) GetSilenceByName(ctx context.Context, namespace, silencedID string) (*types.Silenced, error) {
	args := s.Called(ctx, namespace, silencedID)
	return args.Get(0).(*types.Silenced), args.Error(1)
}

func (s *MockStore) GetSilencesByName(ctx context.Context, namespace string, names []string) ([]*types.Silenced, error) {
	args := s.Called(ctx, namespace, names)
	return args.Get(0).([]*types.Silenced), args.Error(1)
}

func (s *MockStore) GetSilencesBySubscription(ctx context.Context, namespace string, subscriptions []string) ([]*types.Silenced, error) {
	args := s.Called(ctx, namespace, subscriptions)
	return args.Get(0).([]*types.Silenced), args.Error(1)
}

func (s *MockStore) GetSilencesByCheck(ctx context.Context, namespace, checkName string) ([]*types.Silenced, error) {
	args := s.Called(ctx, namespace, checkName)
	return args.Get(0).([]*types.Silenced), args.Error(1)
}

func (s *MockStore) UpdateSilence(ctx context.Context, silenced *types.Silenced) error {
	args := s.Called(ctx, silenced)
	return args.Error(0)
}
