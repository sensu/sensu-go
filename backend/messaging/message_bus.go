package messaging

import "github.com/sensu/sensu-go/backend/daemon"

// MessageBus is the interface to the internal messaging system.
type MessageBus interface {
	daemon.Daemon

	// Subscribe allows a consumer to subscribe to a topic,
	// binding a read-only channel to the topic. Topic messages
	// are delivered to the channel as simple byte arrays.
	Subscribe(topic string, consumer string, channel chan<- []byte) error

	// Unsubscribe allows a consumer to unsubscribe from a topic,
	// removing its read-only channel from the topic's bindings.
	// The channel is not closed, as it may still having other
	// topic bindings.
	Unsubscribe(topic string, consumer string) error

	// Publish sends a message to a topic.
	Publish(topic string, message []byte) error
}
