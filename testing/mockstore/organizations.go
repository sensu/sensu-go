package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// DeleteOrganizationByName ...
func (s *MockStore) DeleteOrganizationByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetOrganizations ...
func (s *MockStore) GetOrganizations(ctx context.Context) ([]*types.Organization, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*types.Organization), args.Error(1)
}

// GetOrganizationByName ...
func (s *MockStore) GetOrganizationByName(ctx context.Context, name string) (*types.Organization, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*types.Organization), args.Error(1)
}

// UpdateOrganization ...
func (s *MockStore) UpdateOrganization(ctx context.Context, org *types.Organization) error {
	args := s.Called(ctx, org)
	return args.Error(0)
}
