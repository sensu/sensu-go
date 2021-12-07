package etcd

import (
	"context"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
)

const opampPathPrefix = "opamp"

var opampKeyBuilder = store.NewKeyBuilder(opampPathPrefix)

func (s *Store) UpdateAgentConfig(ctx context.Context, config *corev3.OpampAgentConfig) error {
	return CreateOrUpdate(ctx, s.client, opampKeyBuilder.Build(corev3.OpampAgentConfigResource), "", config)
}

func (s *Store) GetAgentConfig(ctx context.Context) (*corev3.OpampAgentConfig, error) {
	config := &corev3.OpampAgentConfig{}
	err := Get(ctx, s.client, opampKeyBuilder.Build(corev3.OpampAgentConfigResource), config)
	return config, err
}
