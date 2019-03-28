package etcd

import (
	"context"
	"encoding/json"
	"strings"

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
// or the context passed is cancelled, then the channel will be closed. The caller must
// restart the watcher, if needed.
func (s *Store) GetCheckConfigWatcher(ctx context.Context) <-chan store.WatchEventCheckConfig {
	ch := make(chan store.WatchEventCheckConfig)
	watcherChan := s.client.Watch(ctx, checkKeyBuilder.Build(""), clientv3.WithPrefix(), clientv3.WithCreatedNotify())

	go func() {
		defer close(ch)

		for watchResponse := range watcherChan {
			for _, event := range watchResponse.Events {
				var (
					watchEvent  store.WatchEventCheckConfig
					action      store.WatchActionType
					checkConfig *corev2.CheckConfig
				)

				action = GetWatcherAction(event)
				if action == store.WatchUnknown {
					logger.Error("unknown etcd watch action: ", event.Type.String())
				}

				if action == store.WatchDelete {
					key := store.ParseResourceKey(string(event.Kv.Key))
					checkConfig = &corev2.CheckConfig{
						ObjectMeta: corev2.ObjectMeta{
							Namespace: key.Namespace,
							Name:      key.ResourceName,
						},
					}
				} else {
					checkConfig = &corev2.CheckConfig{}
					if err := json.Unmarshal(event.Kv.Value, checkConfig); err != nil {
						logger.WithField("key", event.Kv.Key).WithError(err).Error("unable to unmarshal check config from key")
					}
				}

				watchEvent = store.WatchEventCheckConfig{
					Action:      action,
					CheckConfig: checkConfig,
				}

				ch <- watchEvent
			}
		}
	}()

	return ch
}

// GetEntityWatcher returns a channel that emits WatchEventEntity structs notifying
// the caller that an Entity was updated. If the watcher runs into a terminal error
// or the context passed is cancelled, then the channel will be closed. The caller must
// restart the watcher, if needed.
func (s *Store) GetEntityWatcher(ctx context.Context) <-chan store.WatchEventEntity {
	ch := make(chan store.WatchEventEntity, 1)
	wc := s.client.Watch(ctx, entityKeyBuilder.Build(""), clientv3.WithPrefix(), clientv3.WithCreatedNotify())
	go func() {
		defer close(ch)

		for resp := range wc {
			for _, event := range resp.Events {
				action := GetWatcherAction(event)
				if action == store.WatchUnknown {
					logger.Errorf("unknown etcd watch action: %s", event.Type.String())
					continue
				}

				var entity corev2.Entity
				if action != store.WatchDelete {
					if err := json.Unmarshal(event.Kv.Value, &entity); err != nil {
						logger.WithError(err).Error("error unmarshaling watch event")
						continue
					}
				} else {
					// Fill in the name and namespace of the deleted entity
					parts := strings.Split(string(event.Kv.Key), "/")
					if len(parts) > 1 {
						entity.Name = parts[len(parts)-1]
						entity.Namespace = parts[len(parts)-2]
					}
				}

				watchEvent := store.WatchEventEntity{
					Action: action,
					Entity: &entity,
				}

				ch <- watchEvent
			}
		}
	}()
	return ch
}

// GetTessenConfigWatcher returns a channel that emits WatchEventTessenConfig structs notifying
// the caller that a TessenConfig was updated. If the watcher runs into a terminal error
// or the context passed is cancelled, then the channel will be closed. The caller must
// restart the watcher, if needed.
func (s *Store) GetTessenConfigWatcher(ctx context.Context) <-chan store.WatchEventTessenConfig {
	ch := make(chan store.WatchEventTessenConfig, 1)
	wc := s.client.Watch(ctx, tessenKeyBuilder.Build(""), clientv3.WithPrefix(), clientv3.WithCreatedNotify())
	go func() {
		defer close(ch)

		for resp := range wc {
			for _, event := range resp.Events {
				action := GetWatcherAction(event)
				if action == store.WatchUnknown {
					logger.Errorf("unknown etcd watch action: %s", event.Type.String())
					continue
				}

				var tessen *corev2.TessenConfig
				if action != store.WatchDelete {
					if err := json.Unmarshal(event.Kv.Value, &tessen); err != nil {
						logger.WithError(err).Error("error unmarshaling watch event")
						continue
					}
				} else {
					// use default tessen config if it's deleted
					tessen = corev2.DefaultTessenConfig()
				}

				watchEvent := store.WatchEventTessenConfig{
					Action:       action,
					TessenConfig: tessen,
				}

				ch <- watchEvent
			}
		}
	}()
	return ch
}
