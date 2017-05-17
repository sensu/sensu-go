package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	v3 "github.com/coreos/etcd/clientv3"
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
	cmp := v3.Compare(v3.Version(getUserPath(u.Username)), "=", 0)
	req := v3.OpPut(getUserPath(u.Username), string(userBytes))
	res, err := s.kvc.Txn(context.TODO()).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf("user %s already exists", u.Username)
	}

	return nil
}
