package agentd

import (
	"context"

	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// GetEntityConfigWatcher uses the store to build a watcher for all EntityConfig resources
func GetEntityConfigWatcher(ctx context.Context, store storev2.Interface) <-chan []storev2.WatchEvent {
	return store.GetEntityConfigStore().Watch(ctx, "", "")
}
