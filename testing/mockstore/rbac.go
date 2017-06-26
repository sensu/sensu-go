package mockstore

import "github.com/sensu/sensu-go/types"

// GetRoles ...
func (s *MockStore) GetRoles() ([]*types.Role, error) {
	args := s.Called()
	return args.Get(0).([]*types.Role), args.Error(1)
}

// GetRoleByName ...
func (s *MockStore) GetRoleByName(name string) (*types.Role, error) {
	args := s.Called(name)
	return args.Get(0).(*types.Role), args.Error(1)
}

// CreateRole ...
func (s *MockStore) CreateRole(role *types.Role) error {
	args := s.Called(role)
	return args.Error(0)
}

// UpdateRole ...
func (s *MockStore) UpdateRole(role *types.Role) error {
	args := s.Called(role)
	return args.Error(0)
}

// DeleteRoleByName ...
func (s *MockStore) DeleteRoleByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}
