package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

func getTokenPath(subject, id string) string {
	return fmt.Sprintf("%s/tokens/%s/%s", etcdRoot, subject, id)
}

func (s *etcdStore) CreateToken(claims *types.Claims) error {
	bytes, err := json.Marshal(claims)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getTokenPath(claims.Subject, claims.Id), string(bytes))
	return err
}

// DeleteTokens deletes multiples tokens, belonging to the same subject, with
// a transaction
func (s *etcdStore) DeleteTokens(subject string, ids []string) error {
	if subject == "" || len(ids) == 0 {
		return errors.New("must specify token subject and at least one ID")
	}

	// Construct the list of operations to make in the transaction
	ops := make([]clientv3.Op, len(ids))
	for i := range ops {
		ops[i] = clientv3.OpDelete(getTokenPath(subject, ids[i]))
	}

	res, err := s.kvc.Txn(context.TODO()).Then(ops...).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf("could not delete all tokens")
	}

	return nil
}

func (s *etcdStore) GetToken(subject, id string) (*types.Claims, error) {
	resp, err := s.kvc.Get(context.TODO(), getTokenPath(subject, id), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("token %s for %s does not exist", id, subject)
	}

	claims := &types.Claims{}
	err = json.Unmarshal(resp.Kvs[0].Value, claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
