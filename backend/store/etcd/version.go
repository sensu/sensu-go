package etcd

import (
	"context"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/kvc"
)

const (
	DatabaseVersionKey           = "sensu_database_version"
	EnterpriseDatabaseVersionKey = "sensu_enterprise_database_version"
)

// GetDatabaseVersion gets the current database version.
func GetDatabaseVersion(ctx context.Context, client *clientv3.Client) (int, error) {
	versionPath := path.Join(EtcdRoot, DatabaseVersionKey)
	return getIntegerAtPath(ctx, versionPath, client)
}

// GetEnterpriseDatabaseVersion gets the current enterprise database version.
func GetEnterpriseDatabaseVersion(ctx context.Context, client *clientv3.Client) (int, error) {
	versionPath := path.Join(EtcdRoot, EnterpriseDatabaseVersionKey)
	return getIntegerAtPath(ctx, versionPath, client)
}

func getIntegerAtPath(ctx context.Context, versionPath string, client *clientv3.Client) (int, error) {
	var resp *clientv3.GetResponse
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = client.Get(ctx, versionPath)
		return kvc.RetryRequest(n, err)
	})
	if err != nil {
		return 0, err
	}
	if len(resp.Kvs) == 0 {
		return 0, nil
	}
	var version int
	if _, err := fmt.Sscanf(string(resp.Kvs[0].Value), "%d", &version); err != nil {
		return 0, &store.ErrNotValid{Err: fmt.Errorf("error getting database version: %s", err)}
	}
	return version, nil
}

// SetDatabaseVersion sets the database version.
func SetDatabaseVersion(ctx context.Context, client *clientv3.Client, version int) error {
	versionPath := path.Join(EtcdRoot, DatabaseVersionKey)
	return kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		_, err = client.Put(ctx, versionPath, fmt.Sprintf("%d", version))
		return kvc.RetryRequest(n, err)
	})
}

// SetEnterpriseDatabaseVersion sets the enterprise database version.
func SetEnterpriseDatabaseVersion(ctx context.Context, client *clientv3.Client, version int) error {
	versionPath := path.Join(EtcdRoot, EnterpriseDatabaseVersionKey)
	return kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		_, err = client.Put(ctx, versionPath, fmt.Sprintf("%d", version))
		return kvc.RetryRequest(n, err)
	})
}
