package messaging

// MessageBus is the interface to the internal messaging system.
type MessageBus interface {
	// Subscribe allows a consumer to subscribe to a topic on an optional channel
	// and listen for events on a read-only channel. Events are delivered
	// as simple byte arrays.
	Subscribe(topic, channel string) (<-chan []byte, error)

	// Publish sends a message to a topic over an optional channel. If there's
	// a problem publishing the message, an error is returned.
	Publish(topic, channel string, message []byte) error
}
