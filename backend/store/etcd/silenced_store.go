package etcd

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

// Create new silenced event
// Get silenced events by subscription

const (
	silencedPathPrefix = "silenced"
)

var (
	silencedKeyBuilder = newKeyBuilder(silencedPathPrefix)
)

func getSilencedConfigPath(check *types.CheckConfig) string {
	return silencedKeyBuilder.withResource(check).build(check.Name)
}

func getSilencedPath(ctx context.Context, name string) string {
	return silencedKeyBuilder.withContext(ctx).build(name)
}

// Clear a silenced entry (delete)
func (s *etcdStore) DeleteSilencedByName(ctx context.Context, checkID string) error {
	if checkID == "" {
		return errors.New("must specify checkID")
	}

	_, err := s.kvc.Delete(ctx, getCheckConfigsPath(ctx, checkID))
	return err
}

// Get all silenced events
func (s *etcdStore) GetSilencedEntries(ctx context.Context) ([]*types.Silenced, error) {
	resp, err := s.kvc.Get(ctx, getSilencedPath(ctx, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*types.Silenced{}, nil
	}

	silencedArray := make([]*types.Silenced, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		silencedEntry := &types.Silenced{}
		err = json.Unmarshal(kv.Value, silencedEntry)
		if err != nil {
			return nil, err
		}
		silencedArray[i] = silencedEntry
	}

	return silencedArray, nil
}

// Get individual silenced event by id
func (s *etcdStore) GetSilencedEntry(ctx context.Context, id string) (*types.Silenced, error) {
	if id == "" {
		return nil, errors.New("must specify id")
	}

	resp, err := s.kvc.Get(ctx, getSilencedPath(ctx, id))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	silencedBytes := resp.Kvs[0].Value
	silencedEntry := &types.Silenced{}
	if err := json.Unmarshal(silencedBytes, silencedEntry); err != nil {
		return nil, err
	}

	return silencedEntry, nil
}

// Get all silenced events by subscription
// func (s *etcdStore) GetSilencedEventsBySubscription(ct

// TODO: simplify this function - should pass in sub or check name or make these
// two functions
func (s *etcdStore) GetSilencedEventsBySubscriptionOrCheck(ctx context.Context, checkName, subscription string) ([]*types.Silenced, error) {
	if checkName == "" && subscription == "" {
		return nil, errors.New("must specify a check or subscription")
	} else if checkName == "" {

		resp, err := s.kvc.Get(ctx, getSilencedPath(ctx, checkName))
		if err != nil {
			return []*types.Silenced{}, nil
		}

		silencedArray := make([]*types.Silenced, len(resp.Kvs))
		for i, kv := range resp.Kvs {
			silencedEvent := &types.Silenced{}
			err = json.Unmarshal(kv.Value, silencedEvent)
			if err != nil {
				return nil, err
			}
			silencedArray[i] = silencedEvent
			return silencedArray, nil
		}
	} else {
		resp, err := s.kvc.Get(ctx, getSilencedPath(ctx, subscription))
		if err != nil {

		}
		silencedArray := make([]*types.Silenced, len(resp.Kvs))
		for i, kv := range resp.Kvs {
			silencedEvent := &types.Silenced{}
			err = json.Unmarshal(kv.Value, silencedEvent)
			if err != nil {
				return nil, err
			}
			silencedArray[i] = silencedEvent
			return silencedArray, nil
		}
	}
	return []*types.Silenced{}, nil
}
