package monitor

import (
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
)

// EventCreator creates alert and resolution events given an entity.
type EventCreator interface {
	Warn(entity *types.Entity) error
	Critical(entity *types.Entity) error
	Pass(entity *types.Entity) error
}

// MessageBusEventCreator publishes a message to WizardBus when Alert or Resolve
// is called, formatting the messages approriately based on the entity.
type MessageBusEventCreator struct {
	MessageBus messaging.MessageBus
}

// Can we just pass the expected function type here (createEventType) and call
// it? Or should we move the event type out and just refactor to createEvent and
// pass in the event type (event or keepalive) we want? that might be cleaner.

// Warn sends a check with status of warn for a keepalive.
func (creatorPtr *MessageBusEventCreator) Warn(event *types.Event) error {
	event.Check.Status = 1

	return creatorPtr.MessageBus.Publish(messaging.TopicEventRaw, event)
}

/*
// Critical sends a check with status of critical for a keepalive.
func (creatorPtr *MessageBusEventCreator) Critical(entity *types.Entity) error {
	event := createKeepaliveEvent(entity)
	event.Check.Status = 2

	return creatorPtr.MessageBus.Publish(messaging.TopicEventRaw, event)
}

// Pass sends a check with a status of OK for a keepalive.
func (creatorPtr *MessageBusEventCreator) Pass(entity *types.Entity) error {
	event := createKeepaliveEvent(entity)
	event.Check.Status = 0

	return creatorPtr.MessageBus.Publish(messaging.TopicEventRaw, event)
}

func createEvent(entity *types.Entity) *types.Event {

}

// Move these two functions back into keepalived
func createKeepaliveEvent(entity *types.Entity) *types.Event {
	keepaliveCheck := &types.Check{
		Config: &types.CheckConfig{
			Name:         KeepaliveCheckName,
			Interval:     entity.KeepaliveTimeout,
			Handlers:     []string{KeepaliveHandlerName},
			Environment:  entity.Environment,
			Organization: entity.Organization,
		},
		Status: 1,
	}
	keepaliveEvent := &types.Event{
		Timestamp: time.Now().Unix(),
		Entity:    entity,
		Check:     keepaliveCheck,
	}

	return keepaliveEvent
}

func createRegistrationEvent(entity *types.Entity) *types.Event {
	registrationCheck := &types.Check{
		Config: &types.CheckConfig{
			Name:         RegistrationCheckName,
			Interval:     entity.KeepaliveTimeout,
			Handlers:     []string{RegistrationHandlerName},
			Environment:  entity.Environment,
			Organization: entity.Organization,
		},
		Status: 1,
	}
	registrationEvent := &types.Event{
		Timestamp: time.Now().Unix(),
		Entity:    entity,
		Check:     registrationCheck,
	}

	return registrationEvent
}

*/
