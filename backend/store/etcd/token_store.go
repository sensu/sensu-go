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

func (s *etcdStore) DeleteToken(subject, id string) error {
	if subject == "" || id == "" {
		return errors.New("must specify token subject and ID")
	}

	_, err := s.kvc.Delete(context.TODO(), getTokenPath(subject, id))
	return err
}

func (s *etcdStore) DeleteTokensByUsername(username string) error {
	if username == "" {
		return errors.New("must specify username")
	}

	_, err := s.kvc.Delete(context.TODO(), getTokenPath(username, ""), clientv3.WithPrefix())
	return err
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
