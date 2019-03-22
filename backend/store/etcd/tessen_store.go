package etcd

import (
	"context"
	"fmt"

	v2 "github.com/sensu/sensu-go/api/core/v2"
)

func getTessenPath() string {
	return fmt.Sprintf("%s%s", EtcdRoot, v2.TessenPath)
}

// CreateOrUpdateTessenConfig creates or updates the tessen configuration
func (s *Store) CreateOrUpdateTessenConfig(ctx context.Context, config *v2.TessenConfig) error {
	return CreateOrUpdate(ctx, s.client, getTessenPath(), "", config)
}

// GetTessenConfig gets the tessen configuration
func (s *Store) GetTessenConfig(ctx context.Context) (*v2.TessenConfig, error) {
	config := &v2.TessenConfig{}
	err := Get(ctx, s.client, getTessenPath(), config)
	return config, err
}
