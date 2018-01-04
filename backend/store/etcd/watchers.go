package etcd

import (
	"context"
	"encoding/json"
	"path"

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
func (s *etcdStore) GetCheckConfigWatcher(ctx context.Context) <-chan store.WatchEventCheckConfig {
	ch := make(chan store.WatchEventCheckConfig)

	go func() {
		watcher := clientv3.NewWatcher(s.client)
		path := path.Join(etcdRoot, checksPathPrefix)
		watcherChan := watcher.Watch(ctx, path, clientv3.WithPrefix(), clientv3.WithCreatedNotify())
		defer close(ch)

		var (
			watchEvent  store.WatchEventCheckConfig
			action      store.WatchActionType
			checkConfig *types.CheckConfig
		)

		for watchResponse := range watcherChan {
			for _, event := range watchResponse.Events {
				action = getWatcherAction(event)
				if action == store.WatchUnknown {
					logger.Error("unknown etcd watch action: ", event.Type.String())
				}

				checkConfig = &types.CheckConfig{}
				if err := json.Unmarshal(event.Kv.Value, checkConfig); err != nil {
					logger.WithError(err).Error("unable to unmarshal check config from key: ", event.Kv.Key)
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
func (s *etcdStore) GetAssetWatcher(ctx context.Context) <-chan store.WatchEventAsset {
	ch := make(chan store.WatchEventAsset)

	go func() {
		watcher := clientv3.NewWatcher(s.client)
		path := path.Join(etcdRoot, assetsPathPrefix)
		watcherChan := watcher.Watch(ctx, path, clientv3.WithPrefix(), clientv3.WithCreatedNotify())
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
					logger.WithError(err).Error("unable to unmarshal check config from key: ", event.Kv.Key)
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
