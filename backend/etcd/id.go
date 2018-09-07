package etcd

import (
	"context"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
)

var (
	backendIDKeyPrefix         = store.NewKeyBuilder("backends").Build()
	backendIDLeasePeriod int64 = 60
)

// BackendID creates and returns a new leased backend ID.
func BackendID(ctx context.Context, client *clientv3.Client) (int64, error) {
	// Grant a lease for 60 seconds
	resp, err := client.Grant(ctx, backendIDLeasePeriod)
	if err != nil {
		return 0, fmt.Errorf("error creating backend ID: error granting lease: %s", err)
	}

	// Register the backend's lease - this is for clients that need to be
	// able to send specific backends messages
	value := fmt.Sprintf("%x", resp.ID)
	key := path.Join(backendIDKeyPrefix, value)
	_, err = client.Put(ctx, key, value, clientv3.WithLease(resp.ID))
	if err != nil {
		return 0, fmt.Errorf("error creating backend ID: error creating key: %s", err)
	}

	// Keep the lease alive indefinitely
	ch, err := client.KeepAlive(ctx, resp.ID)
	if err != nil {
		return 0, fmt.Errorf("error creating backend ID: error creating keepalive: %s", err)
	}
	go func() {
		for range ch {
		}
	}()

	return int64(resp.ID), nil
}
