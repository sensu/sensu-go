package agentd

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	etcdstorev2 "github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

// EntityConfigWatcher watches changes to EntityConfig in etcd and publish them
// over the bus as store.WatchEventEntityConfig
func EntityConfigWatcher(ctx context.Context, client *clientv3.Client, bus messaging.MessageBus) {
	key := etcdstorev2.StoreKey(storev2.ResourceRequest{
		Context:   ctx,
		StoreName: new(corev3.EntityConfig).StoreName(),
	})
	w := etcdstore.Watch(ctx, client, key, true)

	go handleResults(w, bus)
}

// handleResults handles the results from the etcd watcher. It uses the
// store.Watcher interface to allow easier testing
func handleResults(w store.Watcher, bus messaging.MessageBus) {
	for response := range w.Result() {
		if response.Type == store.WatchError {
			logger.
				WithError(errors.New(string(response.Object))).
				Error("unexpected error while watching for entity config updates")
			continue
		}

		var (
			configWrapper wrap.Wrapper
			entityConfig  corev3.EntityConfig
		)

		// Decode and unwrap the entity config
		if err := proto.Unmarshal(response.Object, &configWrapper); err != nil {
			fmt.Println(err)
			logger.WithField("key", response.Key).WithError(err).
				Error("unable to unmarshal entity config from key")
			continue
		}
		if err := configWrapper.UnwrapInto(&entityConfig); err != nil {
			logger.WithField("key", response.Key).WithError(err).
				Error("unable to unwrap entity config from key")
			continue
		}

		event := store.WatchEventEntityConfig{
			Action: response.Type,
			Entity: &entityConfig,
		}

		topic := messaging.EntityConfigTopic(entityConfig.Metadata.Namespace, entityConfig.Metadata.Name)
		if err := bus.Publish(topic, &event); err != nil {
			logger.WithField("topic", topic).WithError(err).
				Error("unable to publish an entity config update to the bus")
			continue
		}

		logger.WithField("topic", topic).
			Debug("successfully published an entity config update to the bus")
	}
}
