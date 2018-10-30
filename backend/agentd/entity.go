package agentd

import (
	"context"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// addEntitySubscription appends the entity subscription (using the format
// "entity:entityID") to the subscriptions of an entity
func addEntitySubscription(entityID string, subscriptions []string) []string {
	entityKey := types.GetEntitySubscription(entityID)
	return append(subscriptions, entityKey)
}

// getProxyEntity verifies if a proxy entity id was provided in the given event and if
// so, retrieves the corresponding entity in the store in order to replace the
// event's entity with it. In case no entity exists, we create an entity with
// the proxy class
func getProxyEntity(event *types.Event, s SessionStore) error {
	ctx := context.WithValue(context.Background(), types.NamespaceKey, event.Entity.Namespace)

	// Verify if a proxy entity id, representing a proxy entity, is defined in the check
	if event.HasCheck() && event.Check.ProxyEntityID != "" {
		// Query the store for an entity using the given proxy entity ID
		entity, err := s.GetEntityByID(ctx, event.Check.ProxyEntityID)
		if err != nil {
			return fmt.Errorf("could not query the store for a proxy entity: %s", err)
		}

		// Check if an entity was found for this proxy entity. If not, we need to create it
		if entity == nil {
			entity = &types.Entity{
				ID:            event.Check.ProxyEntityID,
				Class:         types.EntityProxyClass,
				Namespace:     event.Entity.Namespace,
				Subscriptions: addEntitySubscription(event.Check.ProxyEntityID, []string{}),
			}

			if err := s.UpdateEntity(ctx, entity); err != nil {
				return fmt.Errorf("could not create a proxy entity: %s", err)
			}
		}

		event.Entity = entity
	}

	return nil
}
