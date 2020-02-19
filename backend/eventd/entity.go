package eventd

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// createProxyEntity creates a proxy entity for the given event if the entity
// does not exist already and returns the entity created
func createProxyEntity(event *corev2.Event, s store.EntityStore) error {
	entityName := event.Entity.Name

	// Override the entity name with proxy_entity_name if it was provided
	if event.HasCheck() && event.Check.ProxyEntityName != "" {
		entityName = event.Check.ProxyEntityName
	}

	// Determine if the entity exists
	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
	entity, err := s.GetEntityByName(ctx, entityName)
	if err != nil {
		return err
	}

	// If the entity does not exist, create a proxy entity
	if entity == nil {
		if event.Check.ProxyEntityName != "" {
			// Create a brand new entity since we can't rely on the provided entity,
			// which represents the agent's entity
			entity = &corev2.Entity{
				EntityClass:   corev2.EntityProxyClass,
				Subscriptions: []string{corev2.GetEntitySubscription(entityName)},
				ObjectMeta: corev2.ObjectMeta{
					Namespace: event.Entity.Namespace,
					Name:      entityName,
				},
			}
		} else {
			// Use on the provided entity
			entity = event.Entity
			entity.EntityClass = corev2.EntityProxyClass
			entity.Subscriptions = append(entity.Subscriptions, corev2.GetEntitySubscription(entityName))
		}

		if err := s.UpdateEntity(ctx, entity); err != nil {
			return err
		}
	}

	// Replace the event's entity with our entity
	event.Entity = entity
	return nil
}
