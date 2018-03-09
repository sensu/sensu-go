package etcd

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3"
)

func getAuthenticationPath(id string) string {
	return fmt.Sprintf("%s/authentication/%s", EtcdRoot, id)
}

// CreateJWTSecret creates a new JWT secret
func (s *Store) CreateJWTSecret(secret []byte) error {
	// We need to prepare a transaction to verify the version of the key
	// corresponding to the user in etcd in order to ensure we only put the key
	// if it does not exist
	cmp := clientv3.Compare(clientv3.Version(getAuthenticationPath("secret")), "=", 0)
	req := clientv3.OpPut(getAuthenticationPath("secret"), string(secret))
	res, err := s.client.Txn(context.TODO()).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf("a secret already exist")
	}

	return nil
}

// GetJWTSecret retrieves the JWT signing secret
func (s *Store) GetJWTSecret() ([]byte, error) {
	resp, err := s.client.Get(context.TODO(), getAuthenticationPath("secret"), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("secret does not exist")
	}

	return resp.Kvs[0].Value, nil
}

// UpdateJWTSecret replaces the jwt secret with a new one.
func (s *Store) UpdateJWTSecret(secret []byte) error {
	_, err := s.client.Put(context.TODO(), getAuthenticationPath("secret"), string(secret))
	return err
}
