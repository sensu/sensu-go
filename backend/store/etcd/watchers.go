package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

func getWatcherAction(event *clientv3.Event) store.WatchActionType {
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

	go func() {
		watcher := clientv3.NewWatcher(s.client)
		watcherChan := watcher.Watch(ctx, checkKeyBuilder.Build(""), clientv3.WithPrefix(), clientv3.WithCreatedNotify())
		defer close(ch)

		for watchResponse := range watcherChan {
			for _, event := range watchResponse.Events {
				var (
					watchEvent  store.WatchEventCheckConfig
					action      store.WatchActionType
					checkConfig *types.CheckConfig
				)

				action = getWatcherAction(event)
				if action == store.WatchUnknown {
					logger.Error("unknown etcd watch action: ", event.Type.String())
				}

				if action == store.WatchDelete {
					key := string(event.Kv.Key)
					fmt.Println("event's key is: ", key)
					parts := strings.Split(key, "/")
					// TODO(eric): add key splitter
					//  /sensu.io/checks/org/environment/check_name
					// 0/   1    / 2    / 3 / 4         / 5
					checkConfig = &types.CheckConfig{
						Organization: parts[3],
						Environment:  parts[4],
						Name:         parts[5],
					}
				} else {
					checkConfig = &types.CheckConfig{}
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

// GetAssetWatcher returns a channel that emits WatchEventAsset structs notifying
// the caller that an Asset was updated. If the watcher runs into a terminal error
// or the context passed is cancelled, then the channel will be closed. The caller must
// restart the watcher, if needed.
func (s *Store) GetAssetWatcher(ctx context.Context) <-chan store.WatchEventAsset {
	ch := make(chan store.WatchEventAsset)

	go func() {
		watcher := clientv3.NewWatcher(s.client)
		watcherChan := watcher.Watch(ctx, assetKeyBuilder.Build(""), clientv3.WithPrefix(), clientv3.WithCreatedNotify())
		defer close(ch)

		var (
			watchEvent store.WatchEventAsset
			action     store.WatchActionType
			asset      *types.Asset
		)

		for watchResponse := range watcherChan {
			for _, event := range watchResponse.Events {
				action = getWatcherAction(event)
				if action == store.WatchUnknown {
					logger.Error("unknown etcd watch action: ", event.Type.String())
				}

				asset = &types.Asset{}
				if err := json.Unmarshal(event.Kv.Value, asset); err != nil {
					logger.WithField("key", event.Kv.Key).WithError(err).Error("unable to unmarshal check config from key")
				}

				watchEvent = store.WatchEventAsset{
					Action: action,
					Asset:  asset,
				}

				ch <- watchEvent
			}
		}
	}()

	return ch
}

// GetHookConfigWatcher returns a channel that emits WatchEventHookConfig structs notifying
// the caller that a HookConfig was updated. If the watcher runs into a terminal error
// or the context passed is cancelled, then the channel will be closed. The caller must
// restart the watcher, if needed.
func (s *Store) GetHookConfigWatcher(ctx context.Context) <-chan store.WatchEventHookConfig {
	ch := make(chan store.WatchEventHookConfig)

	go func() {
		watcher := clientv3.NewWatcher(s.client)
		watcherChan := watcher.Watch(ctx, hookKeyBuilder.Build(""), clientv3.WithPrefix(), clientv3.WithCreatedNotify())
		defer close(ch)

		var (
			watchEvent store.WatchEventHookConfig
			action     store.WatchActionType
			hookCfg    *types.HookConfig
		)

		for watchResponse := range watcherChan {
			for _, event := range watchResponse.Events {
				action = getWatcherAction(event)
				if action == store.WatchUnknown {
					logger.Error("unknown etcd watch action: ", event.Type.String())
				}

				hookCfg = &types.HookConfig{}
				if err := json.Unmarshal(event.Kv.Value, hookCfg); err != nil {
					logger.WithField("key", event.Kv.Key).WithError(err).Error("unable to unmarshal check config from key")
				}

				watchEvent = store.WatchEventHookConfig{
					Action:     action,
					HookConfig: hookCfg,
				}

				ch <- watchEvent
			}
		}
	}()

	return ch
}
