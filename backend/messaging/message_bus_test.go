package messaging

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscriptionTopic(t *testing.T) {
	expectedTopic := fmt.Sprintf("%s:acme:foo", TopicSubscriptions)
	topic := SubscriptionTopic("acme", "foo")
	assert.Equal(t, expectedTopic, topic)
}
