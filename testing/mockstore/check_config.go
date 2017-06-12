package mockstore

import "github.com/sensu/sensu-go/types"

//// CheckConfigs

// GetCheckConfigs ...
func (s *MockStore) GetCheckConfigs() ([]*types.CheckConfig, error) {
	args := s.Called()
	return args.Get(0).([]*types.CheckConfig), args.Error(1)
}

// GetCheckConfigByName ...
func (s *MockStore) GetCheckConfigByName(name string) (*types.CheckConfig, error) {
	args := s.Called(name)
	return args.Get(0).(*types.CheckConfig), args.Error(1)
}

// DeleteCheckConfigByName ...
func (s *MockStore) DeleteCheckConfigByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// UpdateCheckConfig ...
func (s *MockStore) UpdateCheckConfig(check *types.CheckConfig) error {
	args := s.Called(check)
	return args.Error(0)
}
