package keepalived

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/silenced"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// A Deregisterer provides a mechanism for deregistering entities and
// notifying the rest of the backend when a deregistration occurs.
type Deregisterer interface {
	// Deregister an entity and return an error if there was any problem during the
	// deregistration process.
	Deregister(e *corev2.Entity) error
}

// Deregistration is an adapter for deregistering an entity from the store and
// publishing a deregistration event to WizardBus.
type Deregistration struct {
	EntityStore   storev2.Interface
	EventStore    store.EventStore
	MessageBus    messaging.MessageBus
	SilencedCache cache.Cache
	StoreTimeout  time.Duration
}

// Deregister an entity and all of its associated events.
func (d *Deregistration) Deregister(entity *corev2.Entity) error {
	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, entity.Namespace)
	tctx, cancel := context.WithTimeout(ctx, d.StoreTimeout)
	defer cancel()

	rr := storev2.NewResourceRequestFromV2Resource(tctx, entity)
	if err := d.EntityStore.Delete(rr); err != nil {
		return fmt.Errorf("error deleting entity in store: %s", err)
	}

	events, err := d.EventStore.GetEventsByEntity(ctx, entity.Name, &store.SelectionPredicate{})
	if err != nil {
		return fmt.Errorf("error fetching events for entity: %s", err)
	}

	for _, event := range events {
		if !event.HasCheck() {
			return fmt.Errorf("error deleting event without check")
		}

		tctx, cancel := context.WithTimeout(ctx, d.StoreTimeout)
		defer cancel()
		if err := d.EventStore.DeleteEventByEntityCheck(
			tctx, entity.Name, event.Check.Name,
		); err != nil {
			return fmt.Errorf("error deleting event for entity: %s", err)
		}

		event.Check.Output = "Resolving due to entity deregistering"
		event.Check.Status = 0
		event.Check.History = []corev2.CheckHistory{}

		if err := d.MessageBus.Publish(messaging.TopicEvent, event); err != nil {
			return fmt.Errorf("error publishing deregistration event: %s", err)
		}
	}

	if entity.Deregistration.Handler != "" {
		deregistrationCheck := &corev2.Check{
			ObjectMeta:    corev2.NewObjectMeta("deregistration", entity.Namespace),
			Interval:      1,
			Subscriptions: []string{},
			Command:       "",
			Handlers:      []string{entity.Deregistration.Handler},
			Status:        1,
		}

		id, err := uuid.NewRandom()
		if err != nil {
			return err
		}

		deregistrationEvent := &corev2.Event{
			Entity:    entity,
			Check:     deregistrationCheck,
			ID:        id[:],
			Timestamp: time.Now().Unix(),
		}

		// Add any silenced subscriptions to the event
		silenced.GetSilenced(ctx, deregistrationEvent, d.SilencedCache)
		if len(deregistrationEvent.Check.Silenced) > 0 {
			deregistrationEvent.Check.IsSilenced = true
		}

		return d.MessageBus.Publish(messaging.TopicEvent, deregistrationEvent)
	}

	logger.WithField("entity", entity.GetName()).Info("entity deregistered")
	return nil
}
