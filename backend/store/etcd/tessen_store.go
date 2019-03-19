package etcd

import (
	"context"
	"fmt"

	"github.com/sensu/sensu-go/backend/tessen"
)

func getTessenPath() string {
	return fmt.Sprintf("%s/tessen", EtcdRoot)
}

// CreateOrUpdateTessenConfig creates or updates the tessen configuration
func (s *Store) CreateOrUpdateTessenConfig(ctx context.Context, config *tessen.Config) error {
	return CreateOrUpdate(ctx, s.client, getTessenPath(), "", config)
}

// GetTessenConfig gets the tessen configuration
func (s *Store) GetTessenConfig(ctx context.Context) (*tessen.Config, error) {
	config := &tessen.Config{}
	err := Get(ctx, s.client, getTessenPath(), config)
	return config, err
}
