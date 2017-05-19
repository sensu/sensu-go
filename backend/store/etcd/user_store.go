package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

func getUserPath(id string) string {
	return fmt.Sprintf("%s/users/%s", etcdRoot, id)
}

// CreateUser creates a new user
func (s *etcdStore) CreateUser(u *types.User) error {
	userBytes, err := json.Marshal(u)
	if err != nil {
		return err
	}

	// We need to prepare a transaction to verify the version of the key
	// corresponding to the user in etcd in order to ensure we only put the key
	// if it does not exist
	cmp := clientv3.Compare(clientv3.Version(getUserPath(u.Username)), "=", 0)
	req := clientv3.OpPut(getUserPath(u.Username), string(userBytes))
	res, err := s.kvc.Txn(context.TODO()).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf("user %s already exists", u.Username)
	}

	return nil
}

func (s *etcdStore) GetUser(username string) (*types.User, error) {
	resp, err := s.kvc.Get(context.TODO(), getUserPath(username), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("user %s does not exist", username)
	}

	user := &types.User{}
	err = json.Unmarshal(resp.Kvs[0].Value, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *etcdStore) UpdateUser(user *types.User) error {
	bytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	_, err = s.kvc.Put(context.TODO(), getUserPath(user.Username), string(bytes))
	return err
}
