package mockstore

// GetRoles ...
func (s *MockStore) GetRoles(org string) ([]*types.Role, error) {
	args := s.Called(org)
	return args.Get(0).([]*types.Role), args.Error(1)
}

// GetRole ...
func (s *MockStore) GetRole(org, name string) (*types.Role, error) {
	args := s.Called(org.name)
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
func (s *MockStore) DeleteRoleByName(org, name string) error {
	args := s.Called(org, role)
	return args.Error(0)
}
