package messaging

import "github.com/sensu/sensu-go/backend/daemon"

const (
	// TopicEvent is the topic for events that have been written to Etcd and
	// normalized by eventd.
	TopicEvent = "sensu:event"

	// TopicKeepalive is the topic for keepalive events.
	TopicKeepalive = "sensu:keepalive"

	// TopicEventRaw is the Session -> Eventd channel -- for raw events directly
	// from agents, subscribe to this.
	TopicEventRaw = "sensu:event-raw"

	// TopicSubscriptions is the topic prefix for each subscription
	TopicSubscriptions = "sensu:check"
)

// MessageBus is the interface to the internal messaging system.
//
// The MessageBus is a simple implementation of Event Sourcing where you have
// one or more producers publishing events and multiple consumers receiving
// all of the events produced. We've adopted AMQPs "topic" concept allowing
// the bus to route multiple types of messages.
//
// Consumers should be careful to send buffered channels to the MessageBus in
// the Subscribe() method, as Subscribe attempts a non-blocking send to the
// provided channel. If there is no receiver / or if the receiver is not ready
// for the message, _the message will be lost_. Events published to the bus are
// fanned out linearly to all (i.e. ordered) to all subscribers.
type MessageBus interface {
	daemon.Daemon

	// Subscribe allows a consumer to subscribe to a topic,
	// binding a read-only channel to the topic. Topic messages
	// are delivered to the channel as simple byte arrays.
	Subscribe(topic string, consumer string, channel chan<- interface{}) error

	// Unsubscribe allows a consumer to unsubscribe from a topic,
	// removing its read-only channel from the topic's bindings.
	// The channel is not closed, as it may still having other
	// topic bindings.
	Unsubscribe(topic string, consumer string) error

	// Publish sends a message to a topic.
	Publish(topic string, message interface{}) error
}
