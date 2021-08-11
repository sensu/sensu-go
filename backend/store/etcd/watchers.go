package etcd

import (
	"context"
	"reflect"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"go.etcd.io/etcd/client/v3"
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

			if checkConfig.Scheduler == "" {
				if checkConfig.RoundRobin {
					checkConfig.Scheduler = corev2.EtcdScheduler
				} else {
					checkConfig.Scheduler = corev2.MemoryScheduler
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

// GetResourceV3Watcher returns a channel that emits WatchEventResourceV3
// structs notifying that a corev3.Resource has been created, deleted or
// updated.
func GetResourceV3Watcher(ctx context.Context, client *clientv3.Client, key string) <-chan store.WatchEventResourceV3 {
	w := Watch(ctx, client, key, true)
	c := make(chan store.WatchEventResourceV3, 1)

	go func() {
		defer logger.Info("closed ResourceV3 watcher")
		defer close(c)

		for {
			select {
			case <-ctx.Done():
				return

			case event := <-w.Result():
				logger := logger.WithField("key", event.Key)

				switch action := event.Type; {
				case action == store.WatchCreate, action == store.WatchDelete, action == store.WatchUpdate:
					wrapper := &wrap.Wrapper{}
					if err := wrapper.Unmarshal(event.Object); err != nil {
						logger.WithError(err).Error("unable to unmarshal resource from key")
						continue
					}

					component, err := wrapper.Unwrap()
					if err != nil {
						logger.WithError(err).Error("unable to unwrap wrapper")
						continue
					}

					c <- store.WatchEventResourceV3{
						Action:   action,
						Resource: component,
					}

				case action == store.WatchError:
					logger.Error("error from underlying etcd watcher")
					c <- store.WatchEventResourceV3{
						Action: action,
					}

				case action == store.WatchUnknown:
					logger.Warn("unknown message from underlying etcd watcher")
					c <- store.WatchEventResourceV3{
						Action: action,
					}
				}
			}
		}
	}()

	return c
}
