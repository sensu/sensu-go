package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

func getTokenPath(id string) string {
	return fmt.Sprintf("%s/tokens/%s", etcdRoot, id)
}

func (s *etcdStore) CreateToken(claims *types.Claims) error {
	bytes, err := json.Marshal(claims)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getTokenPath(claims.Id), string(bytes))
	return err
}

func (s *etcdStore) DeleteToken(jti string) error {
	if jti == "" {
		return errors.New("must specify token ID")
	}

	_, err := s.kvc.Delete(context.TODO(), getTokenPath(jti))
	return err
}

func (s *etcdStore) GetToken(jti string) (*types.Claims, error) {
	resp, err := s.kvc.Get(context.TODO(), getTokenPath(jti), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("token %s does not exist", jti)
	}

	claims := &types.Claims{}
	err = json.Unmarshal(resp.Kvs[0].Value, claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
