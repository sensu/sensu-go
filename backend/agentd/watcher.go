package agentd

import (
	"context"
	"errors"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	etcdstorev2 "github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"go.etcd.io/etcd/client/v3"
)

// GetEntityConfigWatcher watches changes to EntityConfig in etcd and publish them
// over the bus as store.WatchEventEntityConfig
func GetEntityConfigWatcher(ctx context.Context, client *clientv3.Client) <-chan store.WatchEventEntityConfig {
	key := etcdstorev2.StoreKey(storev2.ResourceRequest{
		Context:   ctx,
		StoreName: new(corev3.EntityConfig).StoreName(),
	})
	w := etcdstore.Watch(ctx, client, key, true)
	ch := make(chan store.WatchEventEntityConfig, 1)

	go func() {
		defer close(ch)
		for response := range w.Result() {
			if response.Type == store.WatchError {
				logger.
					WithError(errors.New(string(response.Object))).
					Error("unexpected error while watching for entity config updates")
				ch <- store.WatchEventEntityConfig{
					Action: response.Type,
				}
				continue
			}

			var (
				configWrapper wrap.Wrapper
				entityConfig  corev3.EntityConfig
			)

			// Decode and unwrap the entity config
			if err := proto.Unmarshal(response.Object, &configWrapper); err != nil {
				logger.WithField("key", response.Key).WithError(err).
					Error("unable to unmarshal entity config from key")
				continue
			}
			if err := configWrapper.UnwrapInto(&entityConfig); err != nil {
				logger.WithField("key", response.Key).WithError(err).
					Error("unable to unwrap entity config from key")
				continue
			}

			// Remove the managed_by label if the value is sensu-agent, in case the
			// entity is no longer managed by its agent
			if entityConfig.Metadata.Labels[corev2.ManagedByLabel] == "sensu-agent" {
				delete(entityConfig.Metadata.Labels, corev2.ManagedByLabel)
			}

			ch <- store.WatchEventEntityConfig{
				Action: response.Type,
				Entity: &entityConfig,
			}
		}
	}()

	return ch
}
