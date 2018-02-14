package keepalived

import (
	"context"
	"fmt"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// A Deregisterer provides a mechanism for deregistering entities and
// notifying the rest of the backend when a deregistration occurs.
type Deregisterer interface {
	// Deregister an entity and return an error if there was any problem during the
	// deregistration process.
	Deregister(e *types.Entity) error
}

// Deregistration is an adapter for deregistering an entity from the store and
// publishing a deregistration event to WizardBus.
type Deregistration struct {
	Store      store.Store
	MessageBus messaging.MessageBus
}

// Deregister an entity and all of its associated events.
func (adapterPtr *Deregistration) Deregister(entity *types.Entity) error {
	ctx := context.WithValue(context.Background(), types.OrganizationKey, entity.Organization)
	ctx = context.WithValue(ctx, types.EnvironmentKey, entity.Environment)

	if err := adapterPtr.Store.DeleteEntity(ctx, entity); err != nil {
		return fmt.Errorf("error deleting entity in store: %s", err.Error())
	}

	events, err := adapterPtr.Store.GetEventsByEntity(ctx, entity.ID)
	if err != nil {
		return fmt.Errorf("error fetching events for entity: %s", err.Error())
	}

	for _, event := range events {
		if err := adapterPtr.Store.DeleteEventByEntityCheck(
			ctx, entity.ID, event.Check.Name,
		); err != nil {
			return fmt.Errorf("error deleting event for entity: %s", err.Error())
		}

		event.Check.Output = "Resolving due to entity deregistering"
		event.Check.Status = 0
		event.Check.History = []types.CheckHistory{}

		if err := adapterPtr.MessageBus.Publish(messaging.TopicEvent, event); err != nil {
			return fmt.Errorf("error publishing deregistration event: %s", err.Error())
		}
	}

	if entity.Deregistration.Handler != "" {
		deregistrationCheck := &types.Check{
			Name:          "deregistration",
			Interval:      entity.KeepaliveTimeout,
			Subscriptions: []string{""},
			Command:       "",
			Handlers:      []string{entity.Deregistration.Handler},
			Environment:   entity.Environment,
			Organization:  entity.Organization,
			Status:        1,
		}

		deregistrationEvent := &types.Event{
			Entity: entity,
			Check:  deregistrationCheck,
		}

		return adapterPtr.MessageBus.Publish(messaging.TopicEvent, deregistrationEvent)
	}

	return nil
}
