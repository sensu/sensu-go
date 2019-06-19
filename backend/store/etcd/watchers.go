package etcd

import (
	"context"
	"reflect"

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
				if err := unmarshal(response.Object, &checkConfig); err != nil {
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

// GetTessenConfigWatcher returns a channel that emits WatchEventTessenConfig
// structs notifying the caller that a TessenConfig was updated. If the watcher
// runs into a terminal error or the context passed is cancelled, then the
// channel will be closed.
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
				if err := unmarshal(response.Object, &tessen); err != nil {
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

// GetResourceWatcher ...
func GetResourceWatcher(ctx context.Context, client *clientv3.Client, key string, elemType reflect.Type) <-chan store.WatchEventResource {
	w := Watch(ctx, client, key, true)
	ch := make(chan store.WatchEventResource, 1)

	go func() {
		defer close(ch)
		for response := range w.Result() {
			if response.Type == store.WatchUnknown {
				logger.Error("unknown etcd watch type: ", response.Type)
				continue
			}

			var resource corev2.Resource
			elemPtr := reflect.New(elemType.Elem())

			if response.Type == store.WatchDelete {
				key := store.ParseResourceKey(response.Key)

				meta := elemPtr.Elem().FieldByName("ObjectMeta")
				if !meta.CanSet() {
					logger.WithField("key", response.Key).Error("unable to set the resource object meta")
					continue
				}
				if meta.FieldByName("Name").CanSet() {
					meta.FieldByName("Name").SetString(key.ResourceName)
				}
				if meta.FieldByName("Namespace").CanSet() {
					meta.FieldByName("Namespace").SetString(key.Namespace)
				}
			} else {
				if err := unmarshal(response.Object, elemPtr.Interface()); err != nil {
					logger.WithField("key", response.Key).WithError(err).
						Error("unable to unmarshal resource from key")
					continue
				}
			}

			resource = elemPtr.Interface().(corev2.Resource)
			ch <- store.WatchEventResource{
				Action:   response.Type,
				Resource: resource,
			}
		}
	}()

	return ch
}
