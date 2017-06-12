package mockstore

import "github.com/sensu/sensu-go/types"

//// CheckConfigs

// GetCheckConfigs ...
func (s *MockStore) GetCheckConfigs(org string) ([]*types.CheckConfig, error) {
	args := s.Called(org)
	return args.Get(0).([]*types.CheckConfig), args.Error(1)
}

// GetCheckConfigByName ...
func (s *MockStore) GetCheckConfigByName(org, name string) (*types.CheckConfig, error) {
	args := s.Called(org, name)
	return args.Get(0).(*types.CheckConfig), args.Error(1)
}

// DeleteCheckConfigByName ...
func (s *MockStore) DeleteCheckConfigByName(org, name string) error {
	args := s.Called(org, name)
	return args.Error(0)
}

// UpdateCheckConfig ...
func (s *MockStore) UpdateCheckConfig(check *types.CheckConfig) error {
	args := s.Called(check)
	return args.Error(0)
}
