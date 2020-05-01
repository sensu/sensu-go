package etcd

import (
	"context"
	"errors"

	"github.com/coreos/etcd/clientv3"
	"github.com/google/uuid"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	clusterIDPrefix = "cluster_id"
)

var (
	clusterIDKeyBuilder = store.NewKeyBuilder(clusterIDPrefix)
)

// CreateClusterID creates a sensu cluster id
func (s *Store) CreateClusterID(ctx context.Context, id string) error {
	return errors.New("CreateClusterID is deprecated, use GetClusterID only")
}

// GetClusterID gets the sensu cluster id
func (s *Store) GetClusterID(ctx context.Context) (string, error) {
	key := clusterIDKeyBuilder.Build("")
	uid := uuid.New().String()

	getOp := clientv3.OpGet(key, clientv3.WithLimit(1))
	putOp := clientv3.OpPut(key, uid)
	cmp := clientv3.Compare(clientv3.Version(key), "=", 0)

	var resp *clientv3.TxnResponse
	err := Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Txn(ctx).If(cmp).Then(putOp).Else(getOp).Commit()
		return RetryRequest(n, err)
	})
	if err != nil {
		return "", err
	}

	if resp.Succeeded {
		return uid, nil
	}

	getResp := resp.Responses[0].GetResponseRange()
	if len(getResp.Kvs) != 1 {
		return "", &store.ErrInternal{Message: "cluster ID response is empty"}
	}
	return string(getResp.Kvs[0].Value), nil
}
