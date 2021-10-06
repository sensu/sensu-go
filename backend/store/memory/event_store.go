package memory

import (
	"context"
	"errors"
	"path"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/metrics"
	"github.com/sensu/sensu-go/backend/store"
)

type eventPutter interface {
	PutEvent(ctx context.Context, event *corev2.Event) error
}

type EventStore struct {
	db    store.EventStore
	store store.Store
	data  sync.Map
}

type EventStoreConfig struct {
	DB             store.EventStore
	Store          store.Store
	UpdateInterval time.Duration
}

func NewEventStore(ctx context.Context, cfg EventStoreConfig) *EventStore {
	es := &EventStore{
		db:    cfg.DB,
		store: cfg.Store,
	}
	go es.run(ctx, cfg.UpdateInterval)
	return es
}

func (e *EventStore) run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.update(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (e *EventStore) update(ctx context.Context) {
	logger.Info("updating event store")
	count := 0
	e.data.Range(func(keyI, valueI interface{}) bool {
		count++
		value := valueI.([]byte)
		eventBytes, err := snappy.Decode(nil, value)
		if err != nil {
			// unlikely
			logger.WithError(err).Error("error writing event")
			return true
		}
		var event corev2.Event
		if err := proto.Unmarshal(eventBytes, &event); err != nil {
			// unlikely
			logger.WithError(err).Error("error writing event")
			return true
		}
		putter, ok := e.db.(eventPutter)
		if !ok {
			logger.Error("no event putter available!")
			return false
		}
		ctx = store.NamespaceContext(ctx, event.Entity.Namespace)
		if err := putter.PutEvent(ctx, &event); err != nil {
			logger.WithError(err).Error("error writing event")
		}
		return true
	})
	logger.WithField("num_events", count).Info("updated event store")
}

func getEventWithCheckPath(ctx context.Context, entity, check string) (string, error) {
	namespace := corev2.ContextNamespace(ctx)
	if namespace == "" {
		return "", errors.New("namespace missing from context")
	}

	return path.Join(namespace, entity, check), nil
}

func (e *EventStore) DeleteEventByEntityCheck(ctx context.Context, entity, check string) error {
	eventPath, err := getEventWithCheckPath(ctx, entity, check)
	if err != nil {
		return &store.ErrNotValid{Err: err}
	}
	if err := e.db.DeleteEventByEntityCheck(ctx, entity, check); err != nil {
		return err
	}
	_, ok := e.data.LoadAndDelete(eventPath)
	if !ok {
		return &store.ErrNotFound{Key: eventPath}
	}
	return nil
}

func (e *EventStore) GetEvents(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	return e.db.GetEvents(ctx, pred)
}

func (e *EventStore) GetEventsByEntity(ctx context.Context, entity string, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	return e.db.GetEventsByEntity(ctx, entity, pred)
}

func (e *EventStore) GetEventByEntityCheck(ctx context.Context, entity, check string) (*corev2.Event, error) {
	return e.db.GetEventByEntityCheck(ctx, entity, check)
}

// updateCheckHistory takes two events and merges the check result history of
// the second event into the first event.
func updateEventHistory(event *corev2.Event, prevEvent *corev2.Event) error {
	if prevEvent != nil {
		if !prevEvent.HasCheck() {
			return errors.New("invalid previous event")
		}
		event.Check.MergeWith(prevEvent.Check)
	} else {
		// If there was no previous check, we still need to set State and LastOK.
		event.Check.State = corev2.EventFailingState
		if event.Check.Status == 0 {
			event.Check.LastOK = event.Check.Executed
			event.Check.State = corev2.EventPassingState
		}
		event.Check.MergeWith(event.Check)
	}
	return nil
}

func updateOccurrences(check *corev2.Check) {
	if check == nil {
		return
	}

	historyLen := len(check.History)
	if historyLen > 1 && check.History[historyLen-1].Status == check.History[historyLen-2].Status {
		// 1. Occurrences should always be incremented if the current Check status is the same as the previous status (this includes events with the Check status of OK)
		check.Occurrences++
	} else {
		// 2. Occurrences should always reset to 1 if the current Check status is different than the previous status
		check.Occurrences = 1
	}

	if historyLen > 1 && check.History[historyLen-1].Status != 0 && check.History[historyLen-2].Status == 0 {
		// 3. OccurrencesWatermark only resets on the a first non OK Check status (it does not get reset going between warning, critical, unknown)
		check.OccurrencesWatermark = 1
	} else if check.Occurrences <= check.OccurrencesWatermark {
		// 4. OccurrencesWatermark should remain the same when occurrences is less than or equal to the watermark
		return
	} else {
		// 5. OccurrencesWatermark should be incremented if conditions 3 and 4 have not been met.
		check.OccurrencesWatermark++
	}
}

func handleExpireOnResolveEntries(ctx context.Context, event *corev2.Event, st store.Store) error {
	// Make sure we have a check and that the event is a resolution
	if !event.HasCheck() || !event.IsResolution() {
		return nil
	}

	entries, err := st.GetSilencedEntriesByName(ctx, event.Check.Silenced...)
	if err != nil {
		return err
	}
	toDelete := []string{}
	toRetain := []string{}
	for _, entry := range entries {
		if entry.ExpireOnResolve {
			toDelete = append(toDelete, entry.Name)
		} else {
			toRetain = append(toRetain, entry.Name)
		}
	}

	if err := st.DeleteSilencedEntryByName(ctx, toDelete...); err != nil {
		return err
	}
	event.Check.Silenced = toRetain

	return nil
}

func (e *EventStore) UpdateEvent(ctx context.Context, event *corev2.Event) (*corev2.Event, *corev2.Event, error) {
	if event == nil || event.Check == nil {
		return nil, nil, &store.ErrNotValid{Err: errors.New("event has no check")}
	}

	if err := event.Check.Validate(); err != nil {
		return nil, nil, &store.ErrNotValid{Err: err}
	}

	if err := event.Entity.Validate(); err != nil {
		return nil, nil, &store.ErrNotValid{Err: err}
	}

	eventPath, err := getEventWithCheckPath(ctx, event.Entity.Name, event.Check.Name)
	if err != nil {
		return nil, nil, &store.ErrNotValid{Err: err}
	}
	prevCompressedEventBytesIface, ok := e.data.Load(eventPath)
	if !ok {
		// slow path, need to retrieve event from persistent store
		oldEvent, newEvent, err := e.db.UpdateEvent(ctx, event)
		if err != nil {
			return nil, nil, err
		}
		if newEvent == nil {
			newEvent = oldEvent
		}
		newEventBytes, err := proto.Marshal(newEvent)
		if err != nil {
			return nil, nil, &store.ErrEncode{Err: err, Key: eventPath}
		}
		newEventCompressedBytes := snappy.Encode(nil, newEventBytes)
		e.data.Store(eventPath, newEventCompressedBytes)
		return oldEvent, newEvent, nil
	}

	// fast path, we found the event in our store and we can now work with that
	prevCompressedEventBytes := prevCompressedEventBytesIface.([]byte)
	prevEventBytes, err := snappy.Decode(nil, prevCompressedEventBytes)
	if err != nil {
		return nil, nil, &store.ErrDecode{Err: err, Key: eventPath}
	}

	prevEvent := new(corev2.Event)
	if err := proto.Unmarshal(prevEventBytes, prevEvent); err != nil {
		return nil, nil, &store.ErrDecode{Err: err, Key: eventPath}
	}

	if err := updateEventHistory(event, prevEvent); err != nil {
		return nil, nil, &store.ErrNotValid{Err: err}
	}

	updateOccurrences(event.Check)

	persistEvent := event
	typeLabelValue := metrics.EventTypeLabelCheck

	if event.HasMetrics() {
		// Taking pains to not modify our input, set metrics to nil so they are
		// not persisted.
		newEvent := *event
		persistEvent = &newEvent
		persistEvent.Metrics = nil
		typeLabelValue = metrics.EventTypeLabelCheckAndMetrics
	}

	// Truncate check output if the output is larger than MaxOutputSize
	if size := event.Check.MaxOutputSize; size > 0 && int64(len(event.Check.Output)) > size {
		// Taking pains to not modify our input, set a bound on the check
		// output size.
		newEvent := *persistEvent
		persistEvent = &newEvent
		check := *persistEvent.Check
		check.Output = check.Output[:size]
		persistEvent.Check = &check
	}

	if persistEvent.Timestamp == 0 {
		// If the event is being created for the first time, it may not include
		// a timestamp. Use the current time.
		persistEvent.Timestamp = time.Now().Unix()
	}

	// Handle expire on resolve silenced entries
	if err := handleExpireOnResolveEntries(ctx, persistEvent, e.store); err != nil {
		return nil, nil, err
	}

	// update the history
	// marshal the new event and store it.
	eventBytes, err := proto.Marshal(persistEvent)
	if err != nil {
		return nil, nil, &store.ErrEncode{Key: eventPath, Err: err}
	}

	store.EventBytesSummary.WithLabelValues(typeLabelValue).Observe(float64(len(eventBytes)))
	eventCompressedBytes := snappy.Encode(prevCompressedEventBytes, eventBytes)

	e.data.Store(eventPath, eventCompressedBytes)

	return event, prevEvent, nil
}
