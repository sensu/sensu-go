package agentd

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	stringsutil "github.com/sensu/sensu-go/util/strings"
)

// addEntitySubscription appends the entity subscription (using the format
// "entity:entityName") to the subscriptions of an entity
func addEntitySubscription(entityName string, subscriptions []string) []string {
	entitySubscription := corev2.GetEntitySubscription(entityName)

	// Do not add the entity subscription if it already exists
	if stringsutil.InArray(entitySubscription, subscriptions) {
		return subscriptions
	}

	return append(subscriptions, entitySubscription)
}

// createProxyEntity creates a proxy entity for the given event if the entity
// does not exist already and returns the entity created
func createProxyEntity(event *corev2.Event, store SessionStore) error {
	entityName := event.Entity.Name

	// Override the entity name with proxy_entity_name if it was provided
	if event.HasCheck() && event.Check.ProxyEntityName != "" {
		entityName = event.Check.ProxyEntityName
	}

	// Determine if the entity exists
	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
	entity, err := store.GetEntityByName(ctx, entityName)
	if err != nil {
		return err
	}

	// If the entity does not exist, create a proxy entity
	if entity == nil {
		entity = &corev2.Entity{
			EntityClass:   corev2.EntityProxyClass,
			Subscriptions: []string{corev2.GetEntitySubscription(entityName)},
			ObjectMeta: corev2.ObjectMeta{
				Namespace: event.Entity.Namespace,
				Name:      entityName,
			},
		}

		if err := store.UpdateEntity(ctx, entity); err != nil {
			return err
		}
	}

	// Replace the event's entity with our entity
	event.Entity = entity
	return nil
}
