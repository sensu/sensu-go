package agentd

import (
	"context"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

type ErrDeletedEntity string

func (e ErrDeletedEntity) Error() string {
	return fmt.Sprintf("%s was deleted after this event was scheduled, dropping event", string(e))
}

// addEntitySubscription appends the entity subscription (using the format
// "entity:entityName") to the subscriptions of an entity
func addEntitySubscription(entityName string, subscriptions []string) []string {
	entityKey := types.GetEntitySubscription(entityName)
	return append(subscriptions, entityKey)
}

// getProxyEntity verifies if a proxy entity name was provided in the given event and if
// so, retrieves the corresponding entity in the store in order to replace the
// event's entity with it. In case no entity exists, we create an entity with
// the proxy class only if the event doesn't already have an entity attached.
func getProxyEntity(event *types.Event, s SessionStore) error {
	ctx := context.WithValue(context.Background(), types.NamespaceKey, event.Entity.Namespace)

	// Verify if a proxy entity name, representing a proxy entity, is defined in the check
	if event.HasCheck() && event.Check.ProxyEntityName != "" {
		// Query the store for an entity using the given proxy entity name
		entity, err := s.GetEntityByName(ctx, event.Check.ProxyEntityName)
		if err != nil {
			return fmt.Errorf("could not query the store for a proxy entity: %s", err)
		}

		// Check if an entity was found for this proxy entity.
		// If not, we may need to create it.
		// There are a few possible scenarios here:
		//   1: Entity exists in event and in store (normal case, doesn't reach this point)
		//   2: Entity exists in event and not in store (ie: a proxy check was scheduled
		//      for this entity, and the entity was deleted between when the check was scheduled
		//      and when the result was received). We should not create a new entity in this case.
		//   3: Entity does not exist in event or the store (ie: check was not scheduled by
		//      the backend, but was instead submitted via the agent's API). We should
		//      create a new entity in this case.
		if entity == nil {
			// case 2: event has an entity but it doesn't exist in the store anymore
			// because it was deleted.
			if event.Entity != nil {
				return ErrDeletedEntity(event.Check.ProxyEntityName)
			}

			// case 3: event has no entity, and neither does the store. This event was
			// submitted via the agent API, not scheduled by the backend, so we can
			// safely create an entity for it.
			entity = &types.Entity{
				EntityClass:   types.EntityProxyClass,
				Subscriptions: addEntitySubscription(event.Check.ProxyEntityName, []string{}),
				ObjectMeta: types.ObjectMeta{
					Namespace: event.Entity.Namespace,
					Name:      event.Check.ProxyEntityName,
				},
			}

			if err := s.UpdateEntity(ctx, entity); err != nil {
				return fmt.Errorf("could not create a proxy entity: %s", err)
			}
		}

		event.Entity = entity
	}

	return nil
}
