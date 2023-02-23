package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"golang.org/x/time/rate"
)

type memorydb struct {
	limiter   *rate.Limiter
	data      sync.Map
	evictions sync.Map
}

func newMemoryDB(limiter *rate.Limiter) *memorydb {
	return &memorydb{
		limiter: limiter,
	}
}

type eventEntry struct {
	// Serialized and compressed event
	EventBytes []byte

	// Dirty indicates if the entry has yet to be written out to long term storage
	Dirty bool

	// Mu is necessary in the rare instance when multiple goroutines are competing
	// to store different versions of the same event. This is only necessary due to
	// a design flaw in eventd, and could possibly be mitigated there.
	Mu sync.Mutex
}

const eventOutputTruncatedBytesLabel = "sensu.io/output_truncated_bytes"

type SilenceStoreI interface {
	// DeleteSilences deletes all silences matching the given names.
	DeleteSilences(ctx context.Context, namespace string, names []string) error

	// GetSilences returns all silences in the namespace. A nil slice with no error is
	// returned if none were found.
	GetSilences(ctx context.Context, namespace string) ([]*corev2.Silenced, error)

	// GetSilencedsByCheck returns all silences for the given check . A nil
	// slice with no error is returned if none were found.
	GetSilencesByCheck(ctx context.Context, namespace, check string) ([]*corev2.Silenced, error)

	// GetSilencedEntriesBySubscription returns all entries for the given
	// subscription. A nil slice with no error is returned if none were found.
	GetSilencesBySubscription(ctx context.Context, namespace string, subscriptions []string) ([]*corev2.Silenced, error)

	// GetSilenceByName returns an entry using the given namespace and name. An
	// error is returned if the entry is not found.
	GetSilenceByName(ctx context.Context, namespace, name string) (*corev2.Silenced, error)

	// UpdateSilence creates or updates a given silence.
	UpdateSilence(ctx context.Context, silence *corev2.Silenced) error

	// GetSilencesByName gets all the named silence entries.
	GetSilencesByName(ctx context.Context, namespace string, names []string) ([]*corev2.Silenced, error)
}

func (m *memorydb) ReadEntry(namespace, entity, check string) *eventEntry {
	key := strings.Join([]string{namespace, entity, check}, "\n")
	emptyEntry := &eventEntry{}
	result, _ := m.data.LoadOrStore(key, emptyEntry)
	entry := result.(*eventEntry)
	return entry
}

func (m *memorydb) WriteTo(ctx context.Context, s store.EventStore) error {
	var storeErr error
	ctx = store.NoMergeEventContext(ctx)
	m.data.Range(func(key, value any) bool {
		entry := value.(*eventEntry)
		if err := m.writeEntry(ctx, s, key, entry); err != nil {
			storeErr = err
			return false
		}
		return true
	})
	return storeErr
}

func (m *memorydb) writeEntry(ctx context.Context, s store.EventStore, key any, entry *eventEntry) error {
	entry.Mu.Lock()
	defer entry.Mu.Unlock()
	eventBytes, err := snappy.Decode(nil, entry.EventBytes)
	if err != nil {
		// developer error, fatal
		panic(err)
	}
	var event corev2.Event
	if err := proto.Unmarshal(eventBytes, &event); err != nil {
		// developer error, fatal
		panic(err)
	}
	defer func() {
		ekey := strings.Join([]string{event.Entity.Namespace, event.Entity.Name}, "\n")
		_, ok := m.evictions.LoadAndDelete(ekey)
		if ok {
			m.data.Delete(key)
		}
	}()
	if !entry.Dirty {
		return nil
	}
	if err := m.limiter.Wait(ctx); err != nil {
		// context cancellation
		return err
	}
	if _, _, err := s.UpdateEvent(ctx, &event); err != nil {
		return err
	}
	entry.Dirty = false
	return nil
}

func (m *memorydb) NotifyAgentState(state messaging.AgentNotification) {
	key := strings.Join([]string{state.Namespace, state.Name}, "\n")
	if state.Connected {
		// if there is a pending eviction and we connected again, simply
		// remove the eviction to debounce the system.
		m.evictions.Delete(key)
	} else {
		// the agent has disconnected, and so we should purge all its entries
		// from the event cache.
		m.evictions.Store(key, struct{}{})
	}
}

type EventStore struct {
	backingStore      store.EventStore
	flushInterval     time.Duration
	eventWriteLimiter *rate.Limiter
	db                *memorydb
	silenceStore      SilenceStoreI
	flushC            chan struct{}
	agentConnState    messaging.ChanSubscriber
	bus               messaging.MessageBus
}

type EventStoreConfig struct {
	BackingStore    store.EventStore
	FlushInterval   time.Duration
	EventWriteLimit rate.Limit
	SilenceStore    SilenceStoreI
	Bus             messaging.MessageBus
}

func NewEventStore(config EventStoreConfig) *EventStore {
	limiter := rate.NewLimiter(config.EventWriteLimit, int(config.EventWriteLimit)/10)
	return &EventStore{
		backingStore:      config.BackingStore,
		flushInterval:     config.FlushInterval,
		db:                newMemoryDB(limiter),
		silenceStore:      config.SilenceStore,
		flushC:            make(chan struct{}),
		eventWriteLimiter: limiter,
		agentConnState:    make(messaging.ChanSubscriber),
		bus:               config.Bus,
	}
}

func (e *EventStore) Start(ctx context.Context) error {
	sub, err := e.bus.Subscribe(messaging.TopicAgentConnectionState, "eventstore", e.agentConnState)
	if err != nil {
		return err
	}
	go e.writeLoop(ctx, sub)
	return nil
}

func (e *EventStore) writeLoop(ctx context.Context, sub messaging.Subscription) {
	defer sub.Cancel()
	ticker := time.NewTicker(e.flushInterval)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-e.flushC:
			e.db.WriteTo(ctx, e.backingStore)
		case <-ticker.C:
			e.db.WriteTo(ctx, e.backingStore)
		case message := <-e.agentConnState:
			notification := message.(messaging.AgentNotification)
			e.db.NotifyAgentState(notification)
		}
	}
}

// Flush immediately writes out all dirty events
func (e *EventStore) Flush() {
	e.flushC <- struct{}{}
}

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

func (e *EventStore) UpdateEvent(ctx context.Context, event *corev2.Event) (*corev2.Event, *corev2.Event, error) {
	if event.Entity.EntityClass != corev2.EntityAgentClass {
		// no memory storage for events where an agent is not bound to this instance.
		if err := e.eventWriteLimiter.Wait(ctx); err != nil {
			return nil, nil, err
		}
		return e.backingStore.UpdateEvent(ctx, event)
	}
	entry := e.db.ReadEntry(event.Entity.Namespace, event.Entity.Name, event.Check.Name)
	entry.Mu.Lock()
	defer entry.Mu.Unlock()
	var prev *corev2.Event
	if entry.EventBytes == nil {
		var err error
		prev, err = e.backingStore.GetEventByEntityCheck(ctx, event.Entity.Name, event.Check.Name)
		if err != nil {
			if _, ok := err.(*store.ErrNotFound); !ok {
				return nil, nil, err
			}
		}
	} else {
		decompressed, err := snappy.Decode(nil, entry.EventBytes)
		if err != nil {
			// fatal developer error
			panic(err)
		}
		var ev corev2.Event
		if err := proto.Unmarshal(decompressed, &ev); err != nil {
			// fatal developer error
			panic(err)
		}
		prev = &ev
	}

	if err := updateEventHistory(event, prev); err != nil {
		return nil, nil, &store.ErrNotValid{Err: err}
	}

	updateOccurrences(event.Check)

	persistEvent := event

	if event.HasMetrics() {
		// Taking pains to not modify our input, set metrics to nil so they are
		// not persisted.
		newEvent := *event
		persistEvent = &newEvent
		persistEvent.Metrics = nil
	}

	// Truncate check output if the output is larger than MaxOutputSize

	maxSize := event.Check.MaxOutputSize
	if truncated := int64(len(event.Check.Output)) - maxSize; maxSize > 0 && truncated > 0 {
		// Taking pains to not modify our input, set a bound on the check
		// output size.
		newEvent := *persistEvent
		persistEvent = &newEvent
		check := *persistEvent.Check
		check.Output = check.Output[:maxSize]
		persistEvent.Check = &check

		// Be bad and mutate the input event's labels.
		if event.Labels == nil {
			event.Labels = make(map[string]string)
			persistEvent.Labels = event.Labels
		}
		event.Labels[eventOutputTruncatedBytesLabel] = fmt.Sprint(truncated)
	}

	if persistEvent.Timestamp == 0 {
		// If the event is being created for the first time, it may not include
		// a timestamp. Use the current time.
		persistEvent.Timestamp = time.Now().Unix()
	}

	// Handle expire on resolve silenced entries
	if err := handleExpireOnResolveEntries(ctx, persistEvent, e.silenceStore); err != nil {
		return nil, nil, err
	}

	eventBytes, err := proto.Marshal(persistEvent)
	if err != nil {
		// fatal developer error
		panic(err)
	}
	compressedEventBytes := snappy.Encode(nil, eventBytes)
	entry.EventBytes = compressedEventBytes
	entry.Dirty = true

	if event.IsResolution() {
		// Need to persist to storage as a UX concern
		ctx = store.NoMergeEventContext(ctx)
		if err := e.eventWriteLimiter.Wait(ctx); err != nil {
			return nil, nil, err
		}
		if _, _, err := e.backingStore.UpdateEvent(ctx, event); err != nil {
			return nil, nil, err
		}
		entry.Dirty = false
	}

	return persistEvent, prev, nil
}

func handleExpireOnResolveEntries(ctx context.Context, event *corev2.Event, st SilenceStoreI) error {
	// Make sure we have a check and that the event is a resolution
	if !event.HasCheck() || !event.IsResolution() {
		return nil
	}

	entries, err := st.GetSilencesByName(ctx, event.Entity.Namespace, event.Check.Silenced)
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

	if err := st.DeleteSilences(ctx, event.Entity.Namespace, toDelete); err != nil {
		return err
	}
	event.Check.Silenced = toRetain

	return nil
}

func (e *EventStore) GetEventByEntityCheck(ctx context.Context, entity, check string) (*corev2.Event, error) {
	return e.backingStore.GetEventByEntityCheck(ctx, entity, check)
}

func (e *EventStore) GetEvents(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	return e.backingStore.GetEvents(ctx, pred)
}

func (e *EventStore) GetEventsByEntity(ctx context.Context, entity string, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	return e.backingStore.GetEventsByEntity(ctx, entity, pred)
}

func (e *EventStore) DeleteEventByEntityCheck(ctx context.Context, entity, check string) error {
	return e.backingStore.DeleteEventByEntityCheck(ctx, entity, check)
}

func (e *EventStore) CountEvents(ctx context.Context, pred *store.SelectionPredicate) (int64, error) {
	return e.backingStore.CountEvents(ctx, pred)
}

func (e *EventStore) EventStoreSupportsFiltering(ctx context.Context) bool {
	return true
}
