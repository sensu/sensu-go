package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sensu/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

const (
	silencedPathPrefix = "silenced"
)

var (
	silencedKeyBuilder = newKeyBuilder(silencedPathPrefix)
)

// Unknown subscriptions or checkNames are
// stored with a splat in string form. Key lookups for subscriptions with a
// splat as a checkName may return multiple values.

// populates type keyBuilder with org and env info, returns a prefix
func getSilencedPath(ctx context.Context, name string) string {
	return silencedKeyBuilder.withContext(ctx).build(name)
}

// Delete a silenced entry by its id (subscription + checkname)
func (s *etcdStore) DeleteSilencedEntryByID(ctx context.Context, silencedID string) error {
	if silencedID == "" {
		return errors.New("must specify id")
	}

	_, err := s.kvc.Delete(ctx, getSilencedPath(ctx, silencedID))
	return err
}

// Get all silenced entries
func (s *etcdStore) GetSilencedEntries(ctx context.Context) ([]*types.Silenced, error) {
	resp, err := query(ctx, s, getSilencedPath)
	if err != nil {
		return nil, err
	}
	silencedArray, err := s.arraySilencedEntries(resp)
	if err != nil {
		return nil, err
	}
	return silencedArray, nil
}

// Get silenced entries by subscription
func (s *etcdStore) GetSilencedEntriesBySubscription(ctx context.Context, subscription string) ([]*types.Silenced, error) {
	if subscription == "" {
		return nil, errors.New("must specify subscription")
	}
	resp, err := s.kvc.Get(ctx, getSilencedPath(ctx, subscription), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	silencedArray, err := s.arraySilencedEntries(resp)
	if err != nil {
		return nil, err
	}
	return silencedArray, nil
}

// Get silenced entries by checkname
func (s *etcdStore) GetSilencedEntriesByCheckName(ctx context.Context, checkName string) ([]*types.Silenced, error) {
	if checkName == "" {
		return nil, errors.New("must specify check name")
	}
	resp, err := s.kvc.Get(ctx, getSilencedPath(ctx, ""), clientv3.WithPrefix())
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

// Get silenced entry by id
func (s *etcdStore) GetSilencedEntryByID(ctx context.Context, id string) (*types.Silenced, error) {
	if id == "" {
		return nil, errors.New("must specify id")
	}

	resp, err := s.kvc.Get(ctx, getSilencedPath(ctx, id))
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

// Create new silenced entry
func (s *etcdStore) UpdateSilencedEntry(ctx context.Context, silenced *types.Silenced) error {
	if err := silenced.Validate(); err != nil {
		return err
	}

	silencedBytes, err := json.Marshal(silenced)
	if err != nil {
		return err
	}
	var req clientv3.Op
	cmp := clientv3.Compare(clientv3.Version(getEnvironmentsPath(silenced.Organization, silenced.Environment)), ">", 0)
	if silenced.Expire > 0 {
		lease, err := s.client.Grant(ctx, silenced.Expire)
		if err != nil {
			return err
		}
		req = clientv3.OpPut(getSilencedPath(ctx, silenced.ID), string(silencedBytes), clientv3.WithLease(lease.ID))
	} else {
		req = clientv3.OpPut(getSilencedPath(ctx, silenced.ID), string(silencedBytes))
	}
	res, err := s.kvc.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the silenced entry %s in environment %s/%s",
			silenced.ID,
			silenced.Organization,
			silenced.Environment,
		)
	}

	return nil
}

// arraySilencedEntries is a helper function to unmarshal entries from json and return
// them as an array
func (s *etcdStore) arraySilencedEntries(resp *clientv3.GetResponse) ([]*types.Silenced, error) {
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
