package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

func getTokenPath(subject, id string) string {
	return fmt.Sprintf("%s/tokens/%s/%s", EtcdRoot, subject, id)
}

// AllowTokens adds the provided tokens to the JWT access list
func (s *Store) AllowTokens(tokens ...*jwt.Token) error {
	claims := make([]*v2.Claims, len(tokens))

	// Get the claims for each tokens
	for i, token := range tokens {
		if c, ok := token.Claims.(*v2.Claims); ok {
			claims[i] = c
			continue
		}
		return errors.New("could not parse all token claims")
	}

	// Construct the list of operations for this transaction
	ops := make([]clientv3.Op, len(tokens))
	for i := range ops {
		bytes, err := json.Marshal(claims[i])
		if err != nil {
			return err
		}

		ops[i] = clientv3.OpPut(getTokenPath(claims[i].Subject, claims[i].Id), string(bytes))
	}

	res, err := s.client.Txn(context.TODO()).Then(ops...).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf("could not allow all tokens")
	}

	return nil
}

// RevokeTokens removes the provided tokens from the JWT access list
func (s *Store) RevokeTokens(claims ...*v2.Claims) error {
	// Construct the list of operations for this transaction
	ops := make([]clientv3.Op, len(claims))
	for i := range ops {
		ops[i] = clientv3.OpDelete(getTokenPath(claims[i].Subject, claims[i].Id))
	}

	res, err := s.client.Txn(context.TODO()).Then(ops...).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf("could not revoke all tokens")
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
