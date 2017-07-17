package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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

	_, err = s.kvc.Put(context.TODO(), getUserPath(claims.Id), string(bytes))
	return err
}

func (s *etcdStore) DeleteToken(jti string) error {
	if jti == "" {
		return errors.New("must specify token ID")
	}

	_, err := s.kvc.Delete(context.TODO(), getTokenPath(jti))
	return err
}
