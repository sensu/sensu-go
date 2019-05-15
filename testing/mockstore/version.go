package mockstore

import (
	"context"

	"github.com/coreos/etcd/client"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// GetVersion ...
func (s *MockStore) GetVersion(ctx context.Context, client client.Client) (*corev2.Version, error) {
	args := s.Called(ctx, client)
	return args.Get(0).(*corev2.Version), args.Error(1)
}
