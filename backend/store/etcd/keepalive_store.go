package etcd

import (
	"context"
	"fmt"
	"strconv"
)

func getKeepalivePath(entityID string) string {
	return fmt.Sprintf("%s/keepalives/%s", etcdRoot, entityID)
}

func (s *etcdStore) GetKeepalive(entityID string) (int64, error) {
	resp, err := s.client.Get(context.Background(), getKeepalivePath(entityID))
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

func (s *etcdStore) UpdateKeepalive(entityID string, expiration int64) error {
	expirationStr := strconv.FormatInt(expiration, 10)
	_, err := s.client.Put(context.Background(), getKeepalivePath(entityID), expirationStr)
	return err
}
