package adapters

import (
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
)

// EventCreator creates alert and resolution events given an entity.
type EventCreator interface {
	Warn(entity *types.Entity) error
	Critical(entity *types.Entity) error
	Resolve(entity *types.Entity) error
}

// MessageBusEventCreator publishes a message to WizardBus when Alert or Resolve
// is called, formatting the messages approriately based on the entity.
type MessageBusEventCreator struct {
	MessageBus messaging.MessageBus
}

// Warn sends a check with status of warn for a keepalive.
func (creatorPtr *MessageBusEventCreator) Warn(entity *types.Entity) error {
	event := creatorPtr.createEvent(entity)
	event.Check.Status = 1

	return creatorPtr.MessageBus.Publish(messaging.TopicEvent, event)
}

// Critical sends a check with status of critical for a keepalive.
func (creatorPtr *MessageBusEventCreator) Critical(entity *types.Entity) error {
	event := creatorPtr.createEvent(entity)
	event.Check.Status = 2

	return creatorPtr.MessageBus.Publish(messaging.TopicEvent, event)
}

// Resolve sends a check with a status of OK for a keepalive.
func (creatorPtr *MessageBusEventCreator) Resolve(entity *types.Entity) error {
	event := creatorPtr.createEvent(entity)
	event.Check.Status = 0

	return creatorPtr.MessageBus.Publish(messaging.TopicEvent, event)
}

func (creatorPtr *MessageBusEventCreator) createEvent(entity *types.Entity) *types.Event {
	keepaliveCheck := &types.Check{
		Config: &types.CheckConfig{
			Name:          "keepalive",
			Interval:      entity.KeepaliveTimeout,
			Subscriptions: []string{""},
			Command:       "",
			Handlers:      []string{"keepalive"},
			Organization:  entity.Organization,
		},
		Status: 1,
	}
	keepaliveEvent := &types.Event{
		Entity: entity,
		Check:  keepaliveCheck,
	}

	return keepaliveEvent
}
