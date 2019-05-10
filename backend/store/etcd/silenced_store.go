package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	silencedPathPrefix = "silenced"
)

var (
	silencedKeyBuilder = store.NewKeyBuilder(silencedPathPrefix)
)

// Unknown subscriptions or checkNames are
// stored with a splat in string form. Key lookups for subscriptions with a
// splat as a checkName may return multiple values.

// GetSilencedPath gets the path of the silenced store.
// populates type keyBuilder with org and env info, returns a prefix
func GetSilencedPath(ctx context.Context, name string) string {
	return silencedKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteSilencedEntryByName a silenced entry by its name (subscription + checkname)
func (s *Store) DeleteSilencedEntryByName(ctx context.Context, silencedName string) error {
	if silencedName == "" {
		return errors.New("must specify name")
	}

	_, err := s.client.Delete(ctx, GetSilencedPath(ctx, silencedName))
	return err
}

// GetSilencedEntries gets all silenced entries.
func (s *Store) GetSilencedEntries(ctx context.Context) ([]*types.Silenced, error) {
	resp, err := s.client.Get(ctx, GetSilencedPath(ctx, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	silencedArray, err := s.arraySilencedEntries(resp)
	if err != nil {
		return nil, err
	}
	return silencedArray, nil
}

// GetSilencedEntriesBySubscription gets all silenced entries that match a subscription.
func (s *Store) GetSilencedEntriesBySubscription(ctx context.Context, subscription string) ([]*types.Silenced, error) {
	if subscription == "" {
		return nil, errors.New("must specify subscription")
	}
	resp, err := s.client.Get(ctx, GetSilencedPath(ctx, subscription), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	silencedArray, err := s.arraySilencedEntries(resp)
	if err != nil {
		return nil, err
	}
	return silencedArray, nil
}

// GetSilencedEntriesByCheckName gets all silenced entries that match a check name.
func (s *Store) GetSilencedEntriesByCheckName(ctx context.Context, checkName string) ([]*types.Silenced, error) {
	if checkName == "" {
		return nil, errors.New("must specify check name")
	}
	resp, err := s.client.Get(ctx, GetSilencedPath(ctx, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	// iterate through response entries
	// add anything with checkName == entry.Check to an array and return
	silencedArray := []*types.Silenced{}
	for _, kv := range resp.Kvs {
		silencedEntry := &types.Silenced{}
		err := json.Unmarshal(kv.Value, silencedEntry)
		if err != nil {
			return nil, err
		}
		if silencedEntry.Check == checkName {
			silencedArray = append(silencedArray, silencedEntry)
		}
	}

	return silencedArray, nil
}

// GetSilencedEntryByName gets a silenced entry by name.
func (s *Store) GetSilencedEntryByName(ctx context.Context, name string) (*types.Silenced, error) {
	if name == "" {
		return nil, errors.New("must specify name")
	}

	resp, err := s.client.Get(ctx, GetSilencedPath(ctx, name))
	if err != nil {
		return nil, err
	}
	silencedArray, err := s.arraySilencedEntries(resp)
	if err != nil {
		return nil, err
	}
	if len(silencedArray) < 1 {
		return nil, nil
	}

	return silencedArray[0], nil
}

// UpdateSilencedEntry updates a Silenced.
func (s *Store) UpdateSilencedEntry(ctx context.Context, silenced *types.Silenced) error {
	if err := silenced.Validate(); err != nil {
		return err
	}

	silencedBytes, err := json.Marshal(silenced)
	if err != nil {
		return err
	}
	var req clientv3.Op
	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(silenced.Namespace)), ">", 0)
	if silenced.Expire > 0 {
		// add expire time to begin time, that is the ttl for the lease
		var expireTime int64
		// Check begin time against current time to get an offset for the ttl
		currentTime := time.Now().Unix()
		timeDelta := silenced.Begin - currentTime
		// Add the delta to the expire time unless it is negative (begin time is
		// in the past)
		if timeDelta > 0 {
			expireTime = silenced.Expire + timeDelta
		} else {
			expireTime = silenced.Expire
		}

		lease, err := s.client.Grant(ctx, expireTime)
		if err != nil {
			return err
		}

		req = clientv3.OpPut(GetSilencedPath(ctx, silenced.Name), string(silencedBytes), clientv3.WithLease(lease.ID))
	} else {
		req = clientv3.OpPut(GetSilencedPath(ctx, silenced.Name), string(silencedBytes))
	}
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the silenced entry %s in namespace %s",
			silenced.Name,
			silenced.Namespace,
		)
	}

	return nil
}

// arraySilencedEntries is a helper function to unmarshal entries from json and return
// them as an array
func (s *Store) arraySilencedEntries(resp *clientv3.GetResponse) ([]*types.Silenced, error) {
	if len(resp.Kvs) == 0 {
		return []*types.Silenced{}, nil
	}
	silencedArray := make([]*types.Silenced, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		leaseID := clientv3.LeaseID(kv.Lease)
		ttl, err := s.client.TimeToLive(context.TODO(), leaseID)
		if err != nil {
			return nil, err
		}
		silencedEntry := &types.Silenced{}
		err = json.Unmarshal(kv.Value, silencedEntry)
		if err != nil {
			return nil, err
		}
		silencedEntry.Expire = ttl.TTL
		silencedArray[i] = silencedEntry
	}
	return silencedArray, nil
}
