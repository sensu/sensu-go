package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

func getUserPath(id string) string {
	return fmt.Sprintf("%s/users/%s", etcdRoot, id)
}

func (s *etcdStore) UpdateUser(u *types.User) error {
	userBytes, err := json.Marshal(u)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getUserPath(u.Username), string(userBytes))
	return err
}
