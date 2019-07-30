package etcd

import (
	"context"
	"reflect"

	"github.com/coreos/etcd/clientv3"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

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
			// schedulerd does not support a full refresh of the check schedulers
			if response.Type == store.WatchError {
				continue
			}

			var checkConfig corev2.CheckConfig

			if err := unmarshal(response.Object, &checkConfig); err != nil {
				logger.WithField("key", response.Key).WithError(err).Error("unable to unmarshal check config from key")
				continue
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
			// tessend does not support a full refresh of its config
			if response.Type == store.WatchError {
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
			if response.Type == store.WatchError {
				ch <- store.WatchEventResource{
					Action: response.Type,
				}
				continue
			}

			var resource corev2.Resource
			elemPtr := reflect.New(elemType.Elem())

			if err := unmarshal(response.Object, elemPtr.Interface()); err != nil {
				logger.WithField("key", response.Key).WithError(err).
					Error("unable to unmarshal resource from key")
				continue
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
