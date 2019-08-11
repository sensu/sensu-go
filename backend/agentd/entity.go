package agentd

import (
	"context"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// ErrDeleteEntity is returned when the proxy entity attached to an event was deleted
// after the check was scheduled but before the event was received.
type ErrDeletedEntity string

func (e ErrDeletedEntity) Error() string {
	return fmt.Sprintf("%s was deleted after this event was scheduled, dropping event", string(e))
}

// ErrBadProxyEntity is returned when the entity name for an event does not match
// the proxy entity name in the scheduled check.
// Since the proxy entity is looked up directly and then overwrites the entity inside
// the event, this mismatch could otherwise cause events to be attached to unexpected entities.
type ErrBadProxyEntity [2]string

func (e ErrBadProxyEntity) Error() string {
	return fmt.Sprintf("proxy_entity_name %s does not match existing entity in event %s", e[0], e[1])
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
		//   1: Entity exists in event and in store (normal case)
		//   2: Entity exists in event and not in store (ie: a proxy check was scheduled
		//      for this entity, and the entity was deleted between when the check was scheduled
		//      and when the result was received). We should not create a new entity in this case.
		//   3: Entity does not exist in event or the store (ie: check was not scheduled by
		//      the backend, but was instead submitted via the agent's API). We should
		//      create a new entity in this case.
		if entity == nil {
			// case 2: event has an entity but it doesn't exist in the store anymore
			// because it was deleted.
			if event.Entity.Name != "" {
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

		// The entity name is stored in two places in an event:
		//   1. Event.Check.ProxyEntityName (which is used to create the entity var)
		//   2. Event.Entity.Metadata.Name
		//
		// This method looks up the entity in the store based on (1) and then
		// overwrites the existing entity in the event. Before doing this, confirm
		// that (1) and (2) match. If they don't, the event is invalid.
		//
		// Generally, this shouldn't happen because the agent overwrites the
		// proxy_entity_name field with the entity name when it receives events over
		// it's API, but it's an undesirable state of inconsistency that should be
		// avoided regardless.
		if event.Entity.Name != "" && entity.Name != event.Entity.Name {
			return ErrBadProxyEntity([2]string{event.Check.ProxyEntityName, event.Entity.Name})
		}

		event.Entity = entity
	}

	return nil
}
