package agentd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddEntitySubscription(t *testing.T) {
	subscriptions := []string{"subscription"}

	subscriptions = addEntitySubscription("entity1", subscriptions)

	expectedSubscriptions := []string{"subscription", "entity:entity1"}
	assert.Equal(t, expectedSubscriptions, subscriptions)
}
