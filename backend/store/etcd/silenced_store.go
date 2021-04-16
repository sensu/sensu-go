package etcd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/kvc"
	"go.etcd.io/etcd/client/v3"
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

	return kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		_, err = s.client.Txn(ctx).Then(ops...).Commit()
		return kvc.RetryRequest(n, err)
	})
}

// GetSilencedEntries gets all silenced entries.
func (s *Store) GetSilencedEntries(ctx context.Context) ([]*corev2.Silenced, error) {
	var resp *clientv3.GetResponse
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, GetSilencedPath(ctx, ""), clientv3.WithPrefix())
		return kvc.RetryRequest(n, err)
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
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Txn(ctx).Then(ops...).Commit()
		return kvc.RetryRequest(n, err)
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
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, GetSilencedPath(ctx, ""), clientv3.WithPrefix())
		return kvc.RetryRequest(n, err)
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
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, GetSilencedPath(ctx, name))
		return kvc.RetryRequest(n, err)
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
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Txn(ctx).Then(ops...).Commit()
		return kvc.RetryRequest(n, err)
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

	if silenced.ExpireAt == 0 && silenced.Expire > 0 {
		start := time.Now()
		if silenced.Begin > 0 {
			start = time.Unix(silenced.Begin, 0)
		}
		silenced.ExpireAt = start.Add(time.Duration(silenced.Expire) * time.Second).Unix()
	}

	silencedBytes, err := proto.Marshal(silenced)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}
	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(silenced.Namespace)), ">", 0)
	req := clientv3.OpPut(GetSilencedPath(ctx, silenced.Name), string(silencedBytes))
	var res *clientv3.TxnResponse
	err = kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		res, err = s.client.Txn(ctx).If(cmp).Then(req).Commit()
		return kvc.RetryRequest(n, err)
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
//
// if all is true, then arraySilencedEntries will also return expired entries
func (s *Store) arraySilencedEntries(ctx context.Context, resp *clientv3.GetResponse) ([]*corev2.Silenced, error) {
	if len(resp.Kvs) == 0 {
		return []*corev2.Silenced{}, nil
	}
	result := make([]*corev2.Silenced, 0, len(resp.Kvs))
	rejects := []string{}
	for _, kv := range resp.Kvs {
		silenced := &corev2.Silenced{}
		if err := unmarshal(kv.Value, silenced); err != nil {
			return nil, &store.ErrDecode{Err: err}
		}

		var expire int64
		leaseID := clientv3.LeaseID(kv.Lease)
		if leaseID > 0 {
			// legacy expiry mechanism
			ttl, err := s.client.TimeToLive(ctx, leaseID)
			if err != nil {
				logger.WithError(err).Error("error setting TTL on silenced")
				continue
			}
			silenced.ExpireAt = time.Now().Unix() + ttl.TTL

			result = append(result, silenced)
		} else if silenced.ExpireAt > 0 {
			// new expiry mechanism
			expire = int64(time.Until(time.Unix(silenced.ExpireAt, 0)) / time.Second)
			if expire > 0 {
				result = append(result, silenced)
			} else {
				rejects = append(rejects, silenced.Name)
			}
		} else {
			// the silenced entry has no expiration
			silenced.Expire = -1
			result = append(result, silenced)
		}
	}
	if len(rejects) > 0 {
		logger.Infof("deleting %d expired silenced entries", len(rejects))
		if err := s.DeleteSilencedEntryByName(ctx, rejects...); err != nil {
			logger.WithError(err).Error("error deleting expired silenced entries")
		}
	}
	return result, nil
}

func (s *Store) arrayTxnSilencedEntries(ctx context.Context, resp *clientv3.TxnResponse) ([]*corev2.Silenced, error) {
	results := []*corev2.Silenced{}
	rejects := []string{}
	for _, resp := range resp.Responses {
		for _, kv := range resp.GetResponseRange().Kvs {
			var silenced corev2.Silenced
			if err := unmarshal(kv.Value, &silenced); err != nil {
				return nil, &store.ErrDecode{Err: fmt.Errorf("couldn't get silenced entries: %s", err)}
			}

			var expire int64
			leaseID := clientv3.LeaseID(kv.Lease)
			if leaseID > 0 {
				// legacy expiry mechanism
				ttl, err := s.client.TimeToLive(ctx, leaseID)
				if err != nil {
					logger.WithError(err).Error("error setting TTL on silenced")
					continue
				}
				silenced.ExpireAt = time.Now().Unix() + ttl.TTL

				results = append(results, &silenced)
			} else if silenced.ExpireAt > 0 {
				// new expiry mechanism
				expire = int64(time.Until(time.Unix(silenced.ExpireAt, 0)) / time.Second)
				if expire > 0 {
					results = append(results, &silenced)
				} else {
					rejects = append(rejects, silenced.Name)
				}
			} else {
				// the silenced entry has no expiration
				silenced.Expire = -1
				results = append(results, &silenced)
			}
		}
	}
	if len(rejects) > 0 {
		logger.Infof("deleting %d expired silenced entries", len(rejects))
		if err := s.DeleteSilencedEntryByName(ctx, rejects...); err != nil {
			logger.WithError(err).Error("error deleting expired silenced entries")
		}
	}
	return results, nil
}
