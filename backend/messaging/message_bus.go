// Package messaging provides the means of coordination between the different
// components of the Sensu backend.
package messaging

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sensu/sensu-go/backend/daemon"
)

const (
	// TopicEntityConfig is the topic for the entity configuration sent by agentd
	// to agents
	TopicEntityConfig = "sensu:entity-config"

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

	// TopicTessen is the topic prefix for tessen api events to Tessend.
	TopicTessen = "sensu:tessen"

	// TopicTessenMetric is the topic prefix for tessen api metrics to Tessend.
	TopicTessenMetric = "sensu:tessen-metric"

	// TopicKeepaliveRaw is a separate channel for keepalives that
	// allows eventd to process keepalives at a higher priority than
	// regular events.
	TopicKeepaliveRaw = "sensu:keepalive-raw"
)

var (
	topicCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sensu_go_wizard_bus",
			Help: "Number of elements in a topic",
		},
		[]string{"topic"},
	)
)

// A Subscriber receives messages via a channel.
type Subscriber interface {

	// Receiver returns the channel a subscriber uses to receive messages.
	Receiver() chan<- interface{}
}

// A Subscription is a cancellable subscription to a WizardTopic.
type Subscription struct {
	id     string
	cancel func(string) error
}

// Cancel a WizardSubscription.
func (t Subscription) Cancel() error {
	return t.cancel(t.id)
}

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
	// binding a specific Subscriber to the topic. Topic messages
	// are delivered to the subscriber as type `interface{}`.
	Subscribe(topic string, consumer string, subscriber Subscriber) (Subscription, error)

	// Publish sends a message to a topic.
	Publish(topic string, message interface{}) error
}

// EntityConfigTopic is a helper to determine the proper topic name for an
// entity
func EntityConfigTopic(namespace, name string) string {
	return fmt.Sprintf("%s:%s:%s", TopicEntityConfig, namespace, name)
}

// SubscriptionTopic is a helper to determine the proper topic name for a
// subscription based on the namespace
func SubscriptionTopic(namespace, sub string) string {
	return fmt.Sprintf("%s:%s:%s", TopicSubscriptions, namespace, sub)
}
