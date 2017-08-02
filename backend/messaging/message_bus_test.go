package messaging

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscriptionTopic(t *testing.T) {
	expectedTopic := fmt.Sprintf("%s:default:dev:foo", TopicSubscriptions)
	topic := SubscriptionTopic("default", "dev", "foo")
	assert.Equal(t, expectedTopic, topic)
}
