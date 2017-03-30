package messaging

import "github.com/sensu/sensu-go/backend/daemon"

// MessageBus is the interface to the internal messaging system.
type MessageBus interface {
	daemon.Daemon

	// Subscribe allows a consumer to subscribe to a topic on an optional channel
	// and listen for events on a read-only channel. Events are delivered
	// as simple byte arrays.
	Subscribe(topic string, channel chan<- []byte) error

	// Publish sends a message to a topic over an optional channel. If there's
	// a problem publishing the message, an error is returned.
	Publish(topic string, message []byte) error
}
