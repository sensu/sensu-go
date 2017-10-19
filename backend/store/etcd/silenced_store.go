package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
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

// Clear a silenced entry (delete)
// A silence entry can be cleared by specifying the id, or the intersection of
// subscription and or handler to which the entry applies
// check, subscription, or id required
func (s *etcdStore) DeleteSilencedEntry(ctx context.Context, silencedID string) error {
	if silencedID == "" {
		return errors.New("must specify id")
	}

	_, err := s.kvc.Delete(ctx, getCheckConfigsPath(ctx, silencedID))
	return err
}

// Get all silenced events
func (s *etcdStore) GetSilencedEntries(ctx context.Context) ([]*types.Silenced, error) {
	resp, err := s.kvc.Get(ctx, getSilencedPath(ctx, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	silencedArray, err := arrayEntries(resp)
	if err != nil {
		return nil, err
	}
	return silencedArray, nil
}

// Get silenced events by subscription
func (s *etcdStore) GetSilencedEntriesBySubscription(ctx context.Context, subscription string) ([]*types.Silenced, error) {
	if subscription == "" {
		return nil, errors.New("must specify subscription")
	}
	resp, err := s.kvc.Get(ctx, getSilencedPath(ctx, subscription), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	silencedArray, err := arrayEntries(resp)
	if err != nil {
		return nil, err
	}
	return silencedArray, nil
}

// Get silenced events by checkname
func (s *etcdStore) GetSilencedEntriesByCheckName(ctx context.Context, checkName string) ([]*types.Silenced, error) {
	if checkName == "" {
		return nil, errors.New("must specify check name")
	}
	resp, err := s.kvc.Get(ctx, getSilencedPath(ctx, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	// iterate through response entries
	// add anything with checkName == entry.CheckName to an array and return
	silencedArray := make([]*types.Silenced, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		silencedEntry := &types.Silenced{}
		err := json.Unmarshal(kv.Value, silencedEntry)
		if err != nil {
			return nil, err
		}
		if silencedEntry.CheckName == checkName {
			silencedArray[i] = silencedEntry
		}
	}

	return silencedArray, nil
}

// Get silenced event by id
func (s *etcdStore) GetSilencedEntryByID(ctx context.Context, id string) ([]*types.Silenced, error) {
	if id == "" {
		return nil, errors.New("must specify id")
	}

	resp, err := s.kvc.Get(ctx, getSilencedPath(ctx, id))
	if err != nil {
		return nil, err
	}
	silencedArray, err := arrayEntries(resp)
	if err != nil {
		return nil, err
	}

	return silencedArray, nil
}

// Create new silenced event
func (s *etcdStore) UpdateSilencedEntry(ctx context.Context, silenced *types.Silenced) error {
	if err := silenced.Validate(); err != nil {
		fmt.Println(err)
		return err
	}

	checkBytes, err := json.Marshal(silenced)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getEnvironmentsPath(silenced.Organization, silenced.Environment)), ">", 0)
	req := clientv3.OpPut(getSilencedPath(ctx, silenced.ID), string(checkBytes))
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

// arrayEntries is a helper function to unmarshal entries from json and return
// them as an array
func arrayEntries(resp *clientv3.GetResponse) ([]*types.Silenced, error) {
	if len(resp.Kvs) == 0 {
		return []*types.Silenced{}, nil
	}

	silencedArray := make([]*types.Silenced, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		silencedEntry := &types.Silenced{}
		err := json.Unmarshal(kv.Value, silencedEntry)
		if err != nil {
			return nil, err
		}
		silencedArray[i] = silencedEntry
	}
	return silencedArray, nil
}
