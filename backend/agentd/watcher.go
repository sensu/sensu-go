package agentd

import (
	"context"

	corev3 "github.com/sensu/core/v3"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// GetEntityConfigWatcher uses the store to build a watcher for all EntityConfig resources
func GetEntityConfigWatcher(ctx context.Context, store storev2.Interface) <-chan []storev2.WatchEvent {
	tmp := new(corev3.EntityConfig)
	return store.Watch(ctx, storev2.NewResourceRequest(tmp.GetTypeMeta(), "", "", tmp.StoreName()))
}
