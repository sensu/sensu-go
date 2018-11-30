package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

func getTokenPath(subject, id string) string {
	return fmt.Sprintf("%s/tokens/%s/%s", EtcdRoot, subject, id)
}

// CreateToken creates a Claims.
func (s *Store) CreateToken(claims *types.Claims) error {
	bytes, err := json.Marshal(claims)
	if err != nil {
		return err
	}

	_, err = s.client.Put(context.TODO(), getTokenPath(claims.Subject, claims.Id), string(bytes))
	return err
}

// DeleteTokens deletes multiples tokens, belonging to the same subject, with
// a transaction
func (s *Store) DeleteTokens(subject string, ids []string) error {
	if subject == "" || len(ids) == 0 {
		return errors.New("must specify token subject and at least one ID")
	}

	// Construct the list of operations to make in the transaction
	ops := make([]clientv3.Op, len(ids))
	for i := range ops {
		ops[i] = clientv3.OpDelete(getTokenPath(subject, ids[i]))
	}

	res, err := s.client.Txn(context.TODO()).Then(ops...).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf("could not delete all tokens")
	}

	return nil
}

// GetToken gets a Claims.
func (s *Store) GetToken(subject, id string) (*types.Claims, error) {
	key := getTokenPath(subject, id)
	resp, err := s.client.Get(context.TODO(), key, clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, &store.ErrNotFound{Key: key}
	}

	claims := &types.Claims{}
	err = json.Unmarshal(resp.Kvs[0].Value, claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
