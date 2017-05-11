package messaging

const (
	// TopicEvent is the topic for events that have been written to Etcd and
	// normalized by eventd.
	TopicEvent = "sensu:event"

	// TopicKeepalive is the topic for keepalive events.
	TopicKeepalive = "sensu:keepalive"

	// TopicEventRaw is the Session -> Eventd channel -- for raw events directly
	// from agents, subscribe to this.
	TopicEventRaw = "sensu:event-raw"
)

// MessageBus is the interface to the internal messaging system.
type MessageBus interface {
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
