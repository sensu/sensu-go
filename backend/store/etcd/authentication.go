package etcd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	utilbytes "github.com/sensu/sensu-go/util/bytes"
)

func getAuthenticationPath(id string) string {
	return fmt.Sprintf("%s/authentication/%s", EtcdRoot, id)
}

// CreateJWTSecret creates a new JWT secret. DEPRECATED. Returns non-nil error
// in all circumstances. Use UpdateJWTSecret to replace an exist jwt secret.
func (s *Store) CreateJWTSecret(secret []byte) error {
	return errors.New("deprecated method, use GetJWTSecret instead")
}

// GetJWTSecret retrieves the JWT signing secret, creating it if it does not
// already exist.
func (s *Store) GetJWTSecret() ([]byte, error) {
	randomBytes, err := utilbytes.Random(32)
	if err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	opGet := clientv3.OpGet(getAuthenticationPath("secret"), clientv3.WithLimit(1))
	opPut := clientv3.OpPut(getAuthenticationPath("secret"), string(randomBytes))
	cmp := clientv3.Compare(clientv3.Version(getAuthenticationPath("secret")), "=", 0)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	resp, err := s.client.Txn(ctx).If(cmp).Then(opPut).Else(opGet).Commit()
	if err != nil {
		return nil, err
	}
	if resp.Succeeded {
		return randomBytes, nil
	}
	getResp := resp.Responses[0].GetResponseRange()
	if len(getResp.Kvs) != 1 {
		return nil, errors.New("secret response is empty")
	}
	return getResp.Kvs[0].Value, nil
}

// UpdateJWTSecret replaces the jwt secret with a new one.
func (s *Store) UpdateJWTSecret(secret []byte) error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	if _, err := s.client.Put(ctx, getAuthenticationPath("secret"), string(secret)); err != nil {
		return err
	}
	return nil
}
