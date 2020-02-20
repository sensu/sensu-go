package agentd

import (
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
