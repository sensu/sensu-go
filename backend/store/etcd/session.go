package etcd

import (
	"context"
	"fmt"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"go.etcd.io/etcd/client/v3"
)

func userSessionPath(username, sessionID string) string {
	return fmt.Sprintf("%s/user_sessions/%s/%s", EtcdRoot, username, sessionID)
}

// GetSession retrieves the session state uniquely identified by the given
// username and session ID.
func (s *Store) GetSession(ctx context.Context, username, sessionID string) (string, error) {
	sessionPath := userSessionPath(username, sessionID)

	resp, err := s.client.Get(ctx, sessionPath, clientv3.WithLimit(1))
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", &store.ErrNotFound{Key: sessionPath}
	}

	return string(resp.Kvs[0].Value), nil
}

// UpdateSession applies the supplied state to the session uniquely identified
// by the given username and session ID with attached lease and TTl for each key
func (s *Store) UpdateSession(ctx context.Context, username, sessionID, state string) error {

	leaseDuration := jwt.DefaultAccessTokenLifespan
	ttl := int64(leaseDuration.Minutes()+1) * 60
	leaseResp, err := s.client.Grant(ctx, ttl)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	if _, err := s.client.Put(ctx, userSessionPath(username, sessionID), state, clientv3.WithLease(leaseResp.ID)); err != nil {
		return err
	}
	return nil
}

// DeleteSession deletes the session uniquely identified by the given username
// and session ID.
func (s *Store) DeleteSession(ctx context.Context, username, sessionID string) error {
	if _, err := s.client.Delete(ctx, userSessionPath(username, sessionID)); err != nil {
		return err
	}
	return nil
}
