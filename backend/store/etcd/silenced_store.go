package etcd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
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
func GetSilencedPath(ctx context.Context, name string) string {
	return silencedKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteSilencedEntryByName deletes one or more silenced entries by name
func (s *Store) DeleteSilencedEntryByName(ctx context.Context, silencedNames ...string) error {
	if len(silencedNames) == 0 {
		return nil
	}
	ops := make([]clientv3.Op, 0, len(silencedNames))
	for _, silenced := range silencedNames {
		ops = append(ops, clientv3.OpDelete(GetSilencedPath(ctx, silenced)))
	}

	return Backoff(ctx).Retry(func(n int) (done bool, err error) {
		_, err = s.client.Txn(ctx).Then(ops...).Commit()
		return RetryRequest(n, err)
	})
}

// GetSilencedEntries gets all silenced entries.
func (s *Store) GetSilencedEntries(ctx context.Context) ([]*corev2.Silenced, error) {
	var resp *clientv3.GetResponse
	err := Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, GetSilencedPath(ctx, ""), clientv3.WithPrefix())
		return RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}
	silencedArray, err := s.arraySilencedEntries(ctx, resp)
	if err != nil {
		return nil, err
	}
	return silencedArray, nil
}

// GetSilencedEntriesBySubscription gets all silenced entries that match a set of subscriptions.
func (s *Store) GetSilencedEntriesBySubscription(ctx context.Context, subscriptions ...string) ([]*corev2.Silenced, error) {
	if len(subscriptions) == 0 {
		return nil, &store.ErrNotValid{Err: errors.New("couldn't get silenced entries: must specify at least one subscription")}
	}
	var ops []clientv3.Op
	for _, subscription := range subscriptions {
		ops = append(ops, clientv3.OpGet(GetSilencedPath(ctx, subscription), clientv3.WithPrefix()))
	}

	var resp *clientv3.TxnResponse
	err := Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Txn(ctx).Then(ops...).Commit()
		return RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}

	return s.arrayTxnSilencedEntries(ctx, resp)
}

// GetSilencedEntriesByCheckName gets all silenced entries that match a check name.
func (s *Store) GetSilencedEntriesByCheckName(ctx context.Context, checkName string) ([]*corev2.Silenced, error) {
	if checkName == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify check name")}
	}
	var resp *clientv3.GetResponse
	err := Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, GetSilencedPath(ctx, ""), clientv3.WithPrefix())
		return RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}

	// iterate through response entries
	// add anything with checkName == entry.Check to an array and return
	silencedArray := []*corev2.Silenced{}
	for _, kv := range resp.Kvs {
		silencedEntry := &corev2.Silenced{}
		err := unmarshal(kv.Value, silencedEntry)
		if err != nil {
			return nil, &store.ErrDecode{Err: err}
		}
		if silencedEntry.Check == checkName {
			silencedArray = append(silencedArray, silencedEntry)
		}
	}

	return silencedArray, nil
}

// GetSilencedEntryByName gets a silenced entry by name.
func (s *Store) GetSilencedEntryByName(ctx context.Context, name string) (*corev2.Silenced, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	var resp *clientv3.GetResponse
	err := Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, GetSilencedPath(ctx, name))
		return RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}
	silencedArray, err := s.arraySilencedEntries(ctx, resp)
	if err != nil {
		return nil, err
	}
	if len(silencedArray) < 1 {
		return nil, nil
	}

	return silencedArray[0], nil
}

// GetSilencedEntriesByName gets the named silenced entries.
func (s *Store) GetSilencedEntriesByName(ctx context.Context, names ...string) ([]*corev2.Silenced, error) {
	if len(names) == 0 {
		return nil, nil
	}
	ops := make([]clientv3.Op, 0, len(names))
	for _, name := range names {
		ops = append(ops, clientv3.OpGet(GetSilencedPath(ctx, name)))
	}
	var resp *clientv3.TxnResponse
	err := Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Txn(ctx).Then(ops...).Commit()
		return RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}
	return s.arrayTxnSilencedEntries(ctx, resp)
}

// UpdateSilencedEntry updates a Silenced.
func (s *Store) UpdateSilencedEntry(ctx context.Context, silenced *corev2.Silenced) error {
	if err := silenced.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	silencedBytes, err := proto.Marshal(silenced)
	if err != nil {
		return &store.ErrEncode{Err: err}
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

		var lease *clientv3.LeaseGrantResponse
		err := Backoff(ctx).Retry(func(n int) (done bool, err error) {
			lease, err = s.client.Grant(ctx, expireTime)
			return RetryRequest(n, err)
		})
		if err != nil {
			return err
		}

		req = clientv3.OpPut(GetSilencedPath(ctx, silenced.Name), string(silencedBytes), clientv3.WithLease(lease.ID))
	} else {
		req = clientv3.OpPut(GetSilencedPath(ctx, silenced.Name), string(silencedBytes))
	}
	var res *clientv3.TxnResponse
	err = Backoff(ctx).Retry(func(n int) (done bool, err error) {
		res, err = s.client.Txn(ctx).If(cmp).Then(req).Commit()
		return RetryRequest(n, err)
	})
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return &store.ErrNamespaceMissing{Namespace: silenced.Namespace}
	}

	return nil
}

// arraySilencedEntries is a helper function to unmarshal serialized entries and
// return them as an array
func (s *Store) arraySilencedEntries(ctx context.Context, resp *clientv3.GetResponse) ([]*corev2.Silenced, error) {
	if len(resp.Kvs) == 0 {
		return []*corev2.Silenced{}, nil
	}
	silencedArray := make([]*corev2.Silenced, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		leaseID := clientv3.LeaseID(kv.Lease)
		ttl, err := s.client.TimeToLive(ctx, leaseID)
		if err != nil {
			logger.WithError(err).Error("error setting TTL on silenced")
			continue
		}
		silencedEntry := &corev2.Silenced{}
		err = unmarshal(kv.Value, silencedEntry)
		if err != nil {
			return nil, &store.ErrDecode{Err: err}
		}
		silencedEntry.Expire = ttl.TTL
		silencedArray[i] = silencedEntry
	}
	return silencedArray, nil
}

func (s *Store) arrayTxnSilencedEntries(ctx context.Context, resp *clientv3.TxnResponse) ([]*corev2.Silenced, error) {
	results := []*corev2.Silenced{}
	for _, resp := range resp.Responses {
		for _, kv := range resp.GetResponseRange().Kvs {
			leaseID := clientv3.LeaseID(kv.Lease)
			ttl, err := s.client.TimeToLive(ctx, leaseID)
			if err != nil {
				logger.WithError(err).Error("error setting TTL on silenced")
				continue
			}
			var silenced corev2.Silenced
			if err := unmarshal(kv.Value, &silenced); err != nil {
				return nil, &store.ErrDecode{Err: fmt.Errorf("couldn't get silenced entries: %s", err)}
			}
			silenced.Expire = ttl.TTL
			results = append(results, &silenced)
		}
	}
	return results, nil
}
