package mockstore

import (
	"context"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
)

// GetAgentConfig returns the opamp agent config
func (s *MockStore) GetAgentConfig(ctx context.Context) (*corev3.OpampAgentConfig, error) {
	args := s.Called(ctx)
	return args.Get(0).(*corev3.OpampAgentConfig), args.Error(1)
}

// UpdateAgentConfig updates the opamp agent config
func (s *MockStore) UpdateAgentConfig(ctx context.Context, config *corev3.OpampAgentConfig) error {
	args := s.Called(ctx, config)
	return args.Error(0)
}
