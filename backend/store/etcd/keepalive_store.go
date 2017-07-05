package etcd

import (
	"context"
	"path"
	"strconv"

	"github.com/sensu/sensu-go/types"
)

const (
	keepalivesPathPrefix = "keepalives"
)

func getKeepalivePath(ctx context.Context, id string) string {
	var org string

	// Determine the organization
	if value := ctx.Value(types.OrganizationKey); value != nil {
		org = value.(string)
	} else {
		org = ""
	}

	return path.Join(etcdRoot, keepalivesPathPrefix, org, id)
}

func (s *etcdStore) GetKeepalive(ctx context.Context, entityID string) (int64, error) {
	resp, err := s.client.Get(context.Background(), getKeepalivePath(ctx, entityID))
	if err != nil {
		return 0, err
	}

	if len(resp.Kvs) == 0 {
		return 0, nil
	}

	expirationStr := string(resp.Kvs[0].Value)
	expiration, err := strconv.ParseInt(expirationStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return expiration, nil
}

func (s *etcdStore) UpdateKeepalive(ctx context.Context, entityID string, expiration int64) error {
	expirationStr := strconv.FormatInt(expiration, 10)
	_, err := s.client.Put(context.Background(), getKeepalivePath(ctx, entityID), expirationStr)
	return err
}
