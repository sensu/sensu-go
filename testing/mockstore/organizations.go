package mockstore

import "github.com/sensu/sensu-go/types"

// DeleteOrganizationByName ...
func (s *MockStore) DeleteOrganizationByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// GetOrganizations ...
func (s *MockStore) GetOrganizations() ([]*types.Organization, error) {
	args := s.Called()
	return args.Get(0).([]*types.Organization), args.Error(1)
}

// GetOrganizationByName ...
func (s *MockStore) GetOrganizationByName(name string) (*types.Organization, error) {
	args := s.Called(name)
	return args.Get(0).(*types.Organization), args.Error(1)
}

// UpdateOrganization ...
func (s *MockStore) UpdateOrganization(org *types.Organization) error {
	args := s.Called(org)
	return args.Error(0)
}
