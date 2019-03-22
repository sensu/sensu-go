package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

func (s *MockStore) RegisterExtension(ctx context.Context, ext *types.Extension) error {
	args := s.Called(ctx, ext)
	return args.Error(0)
}

func (s *MockStore) DeregisterExtension(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

func (s *MockStore) GetExtension(ctx context.Context, name string) (*types.Extension, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*types.Extension), args.Error(1)
}

func (s *MockStore) GetExtensions(ctx context.Context, pageSize int64, continueToken string) ([]*types.Extension, string, error) {
	args := s.Called(ctx, pageSize, continueToken)
	return args.Get(0).([]*types.Extension), args.String(1), args.Error(2)
}
