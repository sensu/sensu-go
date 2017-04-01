package messaging

import "github.com/sensu/sensu-go/backend/daemon"

// MessageBus is the interface to the internal messaging system.
type MessageBus interface {
	daemon.Daemon

	// Subscribe allows a consumer to subscribe to a topic and channel
	// and listen for events on a read-only channel. Events are delivered
	// as simple byte arrays.
	Subscribe(topic string, consumer string, channel chan<- []byte) error

	// Unsubscribe allows a consumer to subscribe to a topic and channel
	// and listen for events on a read-only channel. Events are delivered
	// as simple byte arrays.
	Unsubscribe(topic string, consumer string) error

	// Publish sends a message to a topic over an optional channel. If there's
	// a problem publishing the message, an error is returned.
	Publish(topic string, message []byte) error
}
