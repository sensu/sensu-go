package messaging

import "github.com/sensu/sensu-go/backend/daemon"

// MessageBus is the interface to the internal messaging system.
type MessageBus interface {
	daemon.Daemon

	// NewSubscriber returns a Subscriber for this MessageBus. It returns a
	// unique identifier for the subscriber and the corresponding channel for
	// events.
	NewSubscriber() (string, <-chan []byte, error)

	// Subscribe allows a subscriber to subscribe to a topic and channel. It
	// returns an error if the subscription already exists.
	Subscribe(topic, subscriberID string) error

	// Unsubscribe removes a subscription for a subscriber id. It returns an
	// error if no subscription exists.
	Unsubscribe(topic, subscriberID string) error

	// Publish sends a message to a topic over an optional channel. If there's
	// a problem publishing the message, an error is returned.
	Publish(topic string, message []byte) error
}
