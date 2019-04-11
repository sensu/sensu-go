package etcd

import (
	"context"
	"encoding/json"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// GetWatcherAction maps an etcd Event to the corresponding WatchActionType.
// This function is exported for use by sensu-enterprise-go's etcd watchers.
func GetWatcherAction(event *clientv3.Event) store.WatchActionType {
	switch event.Type {
	case mvccpb.PUT:
		if event.IsCreate() {
			return store.WatchCreate
		}
		return store.WatchUpdate
	case mvccpb.DELETE:
		return store.WatchDelete
	}

	return store.WatchUnknown
}

// GetCheckConfigWatcher returns a channel that emits WatchEventCheckConfig structs notifying
// the caller that a CheckConfig was updated. If the watcher runs into a terminal error
// or the context passed is cancelled, then the channel will be closed. The
// watcher will do its best to recover on errors.
func (s *Store) GetCheckConfigWatcher(ctx context.Context) <-chan store.WatchEventCheckConfig {
	key := checkKeyBuilder.WithContext(ctx).Build()
	w := Watch(ctx, s.client, key, true)
	ch := make(chan store.WatchEventCheckConfig, 1)

	go func() {
		defer close(ch)
		for response := range w.Result() {
			if response.Type == store.WatchUnknown {
				logger.Error("unknown etcd watch type: ", response.Type)
				continue
			}

			var checkConfig corev2.CheckConfig

			if response.Type == store.WatchDelete {
				meta := store.ParseResourceKey(response.Key)
				checkConfig.Namespace = meta.Namespace
				checkConfig.Name = meta.ResourceName
			} else {
				if err := json.Unmarshal(response.Object, &checkConfig); err != nil {
					logger.WithField("key", response.Key).WithError(err).Error("unable to unmarshal check config from key")
					continue
				}
			}

			ch <- store.WatchEventCheckConfig{
				Action:      response.Type,
				CheckConfig: &checkConfig,
			}
		}
	}()

	return ch
}

// GetEntityWatcher returns a channel that emits WatchEventEntity structs notifying
// the caller that an Entity was updated. If the watcher runs into a terminal error
// or the context passed is cancelled, then the channel will be closed.
// The watcher does its best to recover from errors.
func (s *Store) GetEntityWatcher(ctx context.Context) <-chan store.WatchEventEntity {
	ch := make(chan store.WatchEventEntity, 1)
	key := entityKeyBuilder.WithContext(ctx).Build()
	w := Watch(ctx, s.client, key, true)

	go func() {
		defer close(ch)
		for response := range w.Result() {
			if response.Type == store.WatchUnknown {
				logger.Error("unknown etcd watch type: ", response.Type)
				continue
			}

			var entity corev2.Entity

			if response.Type == store.WatchDelete {
				meta := store.ParseResourceKey(response.Key)
				entity.Namespace = meta.Namespace
				entity.Name = meta.ResourceName
			} else {
				if err := json.Unmarshal(response.Object, &entity); err != nil {
					logger.WithField("key", response.Key).WithError(err).Error("unable to unmarshal entity from key")
					continue
				}
			}

			ch <- store.WatchEventEntity{
				Action: response.Type,
				Entity: &entity,
			}
		}
	}()

	return ch
}

// GetTessenConfigWatcher returns a channel that emits WatchEventTessenConfig structs notifying
// the caller that a TessenConfig was updated. If the watcher runs into a terminal error
// or the context passed is cancelled, then the channel will be closed. The caller must
// The watcher does its best to recover from errors.
func (s *Store) GetTessenConfigWatcher(ctx context.Context) <-chan store.WatchEventTessenConfig {
	ch := make(chan store.WatchEventTessenConfig, 1)
	key := tessenKeyBuilder.WithContext(ctx).Build()
	w := Watch(ctx, s.client, key, false)

	go func() {
		defer close(ch)
		for response := range w.Result() {
			if response.Type == store.WatchUnknown {
				logger.Error("unknown etcd watch type: ", response.Type)
				continue
			}

			var tessen corev2.TessenConfig

			if response.Type == store.WatchDelete {
				tessen = *corev2.DefaultTessenConfig()
			} else {
				if err := json.Unmarshal(response.Object, &tessen); err != nil {
					logger.WithField("key", response.Key).WithError(err).Error("unable to unmarshal tessen config from key")
					continue
				}
			}

			ch <- store.WatchEventTessenConfig{
				Action:       response.Type,
				TessenConfig: &tessen,
			}
		}
	}()

	return ch
}
