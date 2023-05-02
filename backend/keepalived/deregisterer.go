package keepalived

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/silenced"
	"github.com/sensu/sensu-go/backend/store"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// A Deregisterer provides a mechanism for deregistering entities and
// notifying the rest of the backend when a deregistration occurs.
type Deregisterer interface {
	// Deregister an entity and return an error if there was any problem during the
	// deregistration process.
	Deregister(e *corev2.Entity) error
}

type SilencesCache interface {
	Get(namespace string) []cachev2.Value[*corev2.Silenced, corev2.Silenced]
}

// Deregistration is an adapter for deregistering an entity from the store and
// publishing a deregistration event to WizardBus.
type Deregistration struct {
	Store         storev2.Interface
	MessageBus    messaging.MessageBus
	SilencedCache SilencesCache
	StoreTimeout  time.Duration
}

// Deregister an entity and all of its associated events.
func (d *Deregistration) Deregister(entity *corev2.Entity) error {
	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, entity.Namespace)
	tctx, cancel := context.WithTimeout(ctx, d.StoreTimeout)
	defer cancel()

	ecstore := storev2.Of[*corev3.EntityConfig](d.Store)
	if err := ecstore.Delete(tctx, storev2.ID{Namespace: entity.Namespace, Name: entity.Name}); err != nil {
		return fmt.Errorf("error deleting entity in store: %s", err)
	}

	events, err := d.Store.GetEventStore().GetEventsByEntity(ctx, entity.Name, &store.SelectionPredicate{})
	if err != nil {
		return fmt.Errorf("error fetching events for entity: %s", err)
	}

	for _, event := range events {
		if !event.HasCheck() {
			return fmt.Errorf("error deleting event without check")
		}

		tctx, cancel := context.WithTimeout(ctx, d.StoreTimeout)
		defer cancel()
		if err := d.Store.GetEventStore().DeleteEventByEntityCheck(
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

	if entity.Deregistration.Handler != "" || len(entity.Deregistration.Pipelines) > 0 {
		deregistrationCheck := &corev2.Check{
			ObjectMeta:    corev2.NewObjectMeta("deregistration", entity.Namespace),
			Interval:      1,
			Subscriptions: []string{},
			Command:       "",
			Handlers:      []string{entity.Deregistration.Handler},
			Pipelines:     entity.Deregistration.Pipelines,
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
