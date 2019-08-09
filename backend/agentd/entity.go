package agentd

import (
	"context"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

type ErrImpliedEntity string

func (e ErrImpliedEntity) Error() string {
	return fmt.Sprintf("%s does not exist and implied proxy entities are disabled", string(e))
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
// the proxy class only if createMissing is true.
// If createMissing is false and the entity doesn't exist, an ErrImpliedEntity is returned.
func getProxyEntity(event *types.Event, s SessionStore, createMissing bool) error {
	ctx := context.WithValue(context.Background(), types.NamespaceKey, event.Entity.Namespace)

	// Verify if a proxy entity name, representing a proxy entity, is defined in the check
	if event.HasCheck() && event.Check.ProxyEntityName != "" {
		// Query the store for an entity using the given proxy entity name
		entity, err := s.GetEntityByName(ctx, event.Check.ProxyEntityName)
		if err != nil {
			return fmt.Errorf("could not query the store for a proxy entity: %s", err)
		}

		// Handle implied proxy entities
		if entity == nil {
			if !createMissing {
				return ErrImpliedEntity(event.Check.ProxyEntityName)
			}

			// Create a new proxy entity
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
