package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	pgxv4 "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/selector"
	"github.com/sensu/sensu-go/backend/store"
)

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

// Type is the type of a postgres store provider.
const Type = "postgres"

var (
	_ store.EventStore = &EventStore{}
)

type EventStore struct {
	db             *pgxpool.Pool
	silenceStore   SilenceStoreI
	postgresConfig Config
	batcher        *EventBatcher
}

// isFlapping determines if the check is flapping, based on the TotalStateChange
// and configured thresholds
func isFlapping(check *corev2.Check) bool {
	if check == nil {
		return false
	}

	if check.LowFlapThreshold == 0 || check.HighFlapThreshold == 0 {
		return false
	}

	// Is the check already flapping?
	if check.State == corev2.EventFlappingState {
		return check.TotalStateChange > check.LowFlapThreshold
	}

	// The check was not flapping, now determine if it does now
	return check.TotalStateChange >= check.HighFlapThreshold
}

// updateCheckState determines the check state based on whether the check is
// flapping, and its status
func updateCheckState(check *corev2.Check) {
	if check == nil {
		return
	}
	check.TotalStateChange = totalStateChange(check)
	if flapping := isFlapping(check); flapping {
		check.State = corev2.EventFlappingState
	} else if check.Status == 0 {
		check.State = corev2.EventPassingState
		check.LastOK = check.Executed
	} else {
		check.State = corev2.EventFailingState
	}
}

// totalStateChange calculates the total state change percentage for the
// history, which is later used for check state flap detection.
func totalStateChange(check *corev2.Check) uint32 {
	if check == nil || len(check.History) < 21 {
		return 0
	}

	stateChanges := 0.00
	changeWeight := 0.80
	previousStatus := check.History[0].Status

	for i := 1; i <= len(check.History)-1; i++ {
		if check.History[i].Status != previousStatus {
			stateChanges += changeWeight
		}

		changeWeight += 0.02
		previousStatus = check.History[i].Status
	}

	return uint32(float32(stateChanges) / 20 * 100)
}

// NewEventStore creates a NewEventStore. It prepares several queries for
// future use. If there is a non-nil error, it is due to query preparation
// failing.
func NewEventStore(db *pgxpool.Pool, sStore SilenceStoreI, pg Config, producers int) (*EventStore, error) {
	// TODO add these options to postgres.Config
	//workers := pg.BatchWorkers
	//if workers == 0 {
	//	workers = pg.PoolSize * 2
	//	if workers == 0 {
	//		workers = producers
	//	}
	//}
	//if workers > producers {
	//	workers = producers
	//}
	//bufSize := pg.BatchBuffer
	//if bufSize == 0 {
	//	bufSize = DefaultBatchBufferSize
	//}
	//batchSize := pg.BatchSize
	//if batchSize == 0 {
	//	batchSize = 1
	//}
	cfg := BatchConfig{
		BufferSize: DefaultBatchBufferSize,
		BatchSize:  1,
		Consumers:  20,
		Producers:  20,
		DB:         db,
	}
	batcher, err := NewEventBatcher(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create query batcher: %s", err)
	}
	store := &EventStore{
		db:             db,
		silenceStore:   sStore,
		postgresConfig: pg,
		batcher:        batcher,
	}
	return store, nil
}

func getNamespace(ctx context.Context) (string, error) {
	if ns := corev2.ContextNamespace(ctx); ns == "" {
		return "", &store.ErrNotValid{Err: errors.New("namespace missing from context")}
	} else {
		return ns, nil
	}
}

func getHistory(tsArray, statusArray pq.Int64Array, historyIndex int64) ([]corev2.CheckHistory, error) {
	if len(tsArray) != len(statusArray) {
		return nil, &store.ErrNotValid{Err: errors.New("timestamp array and status array lengths differ")}
	}
	if len(tsArray) > 21 {
		return nil, &store.ErrNotValid{Err: errors.New("history too long")}
	}
	result := make([]corev2.CheckHistory, len(tsArray))
	for i := range result {
		result[i].Status = uint32(statusArray[i])
		result[i].Executed = tsArray[i]
	}
	if len(result) == 21 {
		// The maximum history length has been reached and history has wrapped
		// around.
		//
		// Postgres array indices are 1-based, so subtract 1 from historyIndex.
		result = append(result[historyIndex-1:], result[:historyIndex-1]...)
	}
	return result, nil
}

func scanEvents(rows pgxv4.Rows, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	var (
		tsArray, statusArray pq.Int64Array
		historyIndex         int64
		lastOK               int64
		occurrences          int64
		occurrencesWatermark int64
		serialized           []byte
	)
	events := []*corev2.Event{}
	i := int64(0)
	for rows.Next() {
		if err := rows.Scan(&lastOK, &occurrences, &occurrencesWatermark, &tsArray, &statusArray, &historyIndex, &serialized); err != nil {
			return nil, &store.ErrNotValid{Err: fmt.Errorf("error reading events: %s", err)}
		}
		var event corev2.Event
		if err := proto.Unmarshal(serialized, &event); err != nil {
			return nil, &store.ErrDecode{Err: fmt.Errorf("error reading events: %s", err)}
		}
		if event.Check == nil {
			return nil, &store.ErrNotValid{Err: errors.New("nil check")}
		}
		history, err := getHistory(tsArray, statusArray, historyIndex)
		if err != nil {
			return nil, err
		}
		event.Check.History = history
		event.Check.LastOK = lastOK
		event.Check.Occurrences = occurrences
		event.Check.OccurrencesWatermark = occurrencesWatermark
		updateCheckState(event.Check)
		events = append(events, &event)
		i++
	}
	if pred != nil && i < pred.Limit {
		pred.Continue = ""
	}
	if err := rows.Err(); err != nil {
		return nil, &store.ErrInternal{Message: fmt.Sprintf("error reading events: %s", err)}
	}
	return events, nil
}

type continueToken struct {
	Offset int64 `json:"offset"`
}

func (c *continueToken) Encode() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *continueToken) Decode(token string) error {
	if err := json.Unmarshal([]byte(token), c); err != nil {
		return fmt.Errorf("couldn't decode token: %s", err)
	}
	return nil
}

func (e *EventStore) DeleteEventByEntityCheck(ctx context.Context, entity, check string) error {
	ns, err := getNamespace(ctx)
	if err != nil {
		return err
	}
	if entity == "" || check == "" {
		return &store.ErrNotValid{Err: errors.New("must specify entity and check name")}
	}
	if _, err := e.db.Exec(ctx, DeleteEventQuery, ns, entity, check); err != nil {
		return &store.ErrInternal{Message: fmt.Sprintf("couldn't delete event: %s", err)}
	}
	return nil
}

func getLimitAndOffset(pred *store.SelectionPredicate) (sql.NullInt64, int64, error) {
	var limit sql.NullInt64
	var offset int64
	if pred != nil && pred.Limit > 0 {
		limit.Int64, limit.Valid = pred.Limit, true
		var token continueToken
		if pred.Offset > 0 {
			offset = pred.Offset
		}
		if pred.Continue != "" {
			if err := token.Decode(pred.Continue); err != nil {
				return limit, offset, &store.ErrNotValid{Err: fmt.Errorf("couldn't get events: error decoding token: %s", err)}
			}
			offset = token.Offset
		}
		token.Offset += pred.Limit
		pred.Continue = token.Encode()
	}
	return limit, offset, nil
}

func (e *EventStore) GetEvents(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	ns := corev2.ContextNamespace(ctx)
	if ns == corev2.NamespaceTypeAll {
		ns = ""
	}
	query, args, err := CreateGetEventsQuery(ns, "", "", selector.SelectorFromContext(ctx), pred)
	if err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}
	rows, err := e.db.Query(ctx, query, args...)
	if err != nil {
		return nil, &store.ErrInternal{Message: fmt.Sprintf("couldn't get events: %s", err)}
	}
	defer rows.Close()
	return scanEvents(rows, pred)
}

func (e *EventStore) GetEventsByEntity(ctx context.Context, entity string, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	ns, err := getNamespace(ctx)
	if err != nil {
		// Warning: do not wrap this error
		return nil, err
	}
	if entity == "" {
		return nil, &store.ErrNotValid{Err: errors.New("couldn't get events: must specify entity")}
	}
	query, args, err := CreateGetEventsQuery(ns, entity, "", selector.SelectorFromContext(ctx), pred)
	if err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}
	rows, err := e.db.Query(ctx, query, args...)
	if err != nil {
		return nil, &store.ErrInternal{Message: fmt.Sprintf("couldn't get events: %s", err)}
	}
	defer rows.Close()
	return scanEvents(rows, pred)
}

func (e *EventStore) GetEventByEntityCheck(ctx context.Context, entity, check string) (*corev2.Event, error) {
	ns, err := getNamespace(ctx)
	if err != nil {
		// Warning: do not wrap this error
		return nil, err
	}
	if entity == "" || check == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify entity and check name")}
	}
	rows, err := e.db.Query(ctx, GetEventByEntityCheckQuery, ns, entity, check)
	if err != nil {
		return nil, &store.ErrInternal{Message: fmt.Sprintf("couldn't get event: %s", err)}
	}
	defer rows.Close()
	events, err := scanEvents(rows, nil)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}
	return events[0], nil
}

func marshalSelectors(event *corev2.Event) []byte {
	selectors := corev2.EventFields(event)
	for k, v := range event.Labels {
		k = fmt.Sprintf("event.labels.%s", k)
		selectors[k] = v
	}
	if event.HasCheck() {
		for k, v := range event.Check.Labels {
			k = fmt.Sprintf("event.check.labels.%s", k)
			selectors[k] = v
		}
	}
	for k, v := range event.Entity.Labels {
		k = fmt.Sprintf("event.entity.labels.%s", k)
		selectors[k] = v
	}
	b, _ := json.Marshal(selectors)
	return b
}

// UpdateEvent updates the event in the store, returns the fully updated event,
// and the previous event, along with any error encountered.
//
// CONSUMERS BEWARE: The previous event will not be fully spec'd, and is
// mostly returned so that eventd can consult the event.Check.Ttl field.
func (e *EventStore) UpdateEvent(ctx context.Context, event *corev2.Event) (uEvent, pEvent *corev2.Event, eErr error) {
	if event == nil || event.Check == nil {
		return nil, nil, errors.New("event has no check")
	}

	if err := event.Check.Validate(); err != nil {
		return nil, nil, err
	}

	if err := event.Entity.Validate(); err != nil {
		return nil, nil, err
	}

	persistEvent := event

	if event.HasMetrics() {
		// Taking pains to not modify our input, set metrics to nil so they are
		// not persisted. Set metrics back to non-nil before returning the event.
		metrics := event.Metrics
		defer func() {
			if uEvent != nil {
				uEvent.Metrics = metrics
			}
		}()
		newEvent := *event
		persistEvent = &newEvent
		persistEvent.Metrics = nil
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

	ns := persistEvent.Entity.Namespace

	b, err := proto.Marshal(persistEvent)
	if err != nil {
		return nil, nil, &store.ErrEncode{Err: err}
	}

	updateCheckState(event.Check)
	naiveCheckState := event.Check.State
	selectors := marshalSelectors(event)

	args := EventArgs{
		Namespace:  ns,
		Entity:     event.Entity.Name,
		Check:      event.Check.Name,
		LastOK:     event.Check.Executed,
		Status:     int32(event.Check.Status),
		Serialized: b,
		Selectors:  selectors,
	}

	row, err := e.batcher.Do(args)
	if err != nil {
		return nil, nil, err
	}

	if row.LastOK > 0 {
		event.Check.LastOK = row.LastOK
	}
	event.Check.Occurrences = row.Occurrences
	event.Check.OccurrencesWatermark = row.OccurrencesWatermark

	history, err := getHistory(row.HistoryTS, row.HistoryStatus, row.HistoryIndex)
	if err != nil {
		return nil, nil, &store.ErrNotValid{Err: fmt.Errorf("couldn't update event: %s", err)}
	}
	event.Check.History = history

	// Handle expire on resolve silenced entries
	if err = e.handleExpireOnResolveEntries(ctx, event); err != nil {
		return nil, nil, err
	}

	updateCheckState(event.Check)

	if event.Check.State != naiveCheckState {
		// HACK ALERT
		// The check state is flapping, but we didn't know that before getting
		// the history. We'll need to go and update the store again to reflect that.
		// This is an uncommon case, so it's worth losing a little extra performance
		// here in order to make the common case faster and simpler.
		selectors := marshalSelectors(event)
		if _, err := e.db.Exec(ctx, UpdateEventOnlyQuery, b, selectors, event.Entity.Namespace, event.Entity.Name, event.Check.Name); err != nil {
			return nil, nil, &store.ErrInternal{Message: fmt.Sprintf("couldn't update event: %s", err)}
		}
	}

	var prevEvent corev2.Event

	if len(row.PreviousSerialized) > 0 {
		if err := proto.Unmarshal(row.PreviousSerialized, &prevEvent); err != nil {
			return nil, nil, &store.ErrDecode{Err: fmt.Errorf("couldn't update event: %s", err)}
		}
		return event, &prevEvent, nil
	}

	return event, nil, nil
}

func scanCounts(rows pgxv4.Rows) (map[string]EventGauges, error) {
	gauges := map[string]EventGauges{}

	for rows.Next() {
		var (
			namespace      string
			total          int64
			statusOK       int64
			statusWarning  int64
			statusCritical int64
			statusOther    int64
			statePassing   int64
			stateFlapping  int64
			stateFailing   int64
		)

		if err := rows.Scan(&namespace, &total, &statusOK, &statusWarning, &statusCritical, &statusOther, &statePassing, &stateFlapping, &stateFailing); err != nil {
			return nil, &store.ErrNotValid{Err: fmt.Errorf("error reading counts: %s", err)}
		}

		gauges[namespace] = EventGauges{
			Total:          total,
			StatusOK:       statusOK,
			StatusWarning:  statusWarning,
			StatusCritical: statusCritical,
			StatusOther:    statusOther,
			StatePassing:   statePassing,
			StateFlapping:  stateFlapping,
			StateFailing:   stateFailing,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, &store.ErrInternal{Message: fmt.Sprintf("error reading events: %s", err)}
	}

	return gauges, nil
}

// GetEventGaugesByNamespace queries the store and returns a map of EventGauge
// data, indexed by namespace.
func (e *EventStore) GetEventGaugesByNamespace(ctx context.Context) (map[string]EventGauges, error) {
	rows, err := e.db.Query(ctx, GetEventCountsByNamespaceQuery)
	if err != nil {
		return nil, &store.ErrInternal{Message: fmt.Sprintf("couldn't get event gauges: %s", err)}
	}
	defer rows.Close()

	return scanCounts(rows)
}

// GetKeepaliveGaugesByNamespace queries the store and returns a map of
// KeepaliveGauge data, indexed by namespace.
func (e *EventStore) GetKeepaliveGaugesByNamespace(ctx context.Context) (map[string]KeepaliveGauges, error) {
	rows, err := e.db.Query(ctx, GetKeepaliveCountsByNamespaceQuery)
	if err != nil {
		return nil, &store.ErrInternal{Message: fmt.Sprintf("couldn't get keepalive gauges: %s", err)}
	}
	defer rows.Close()

	return scanCounts(rows)
}

// Close closes the underlying db and releases any associated resources.
func (e *EventStore) Close() (err error) {
	e.db.Close()
	return nil
}

func (e *EventStore) handleExpireOnResolveEntries(ctx context.Context, event *corev2.Event) error {
	// Make sure we have a check and that the event is a resolution
	if !event.HasCheck() || !event.IsResolution() || len(event.Check.Silenced) == 0 {
		return nil
	}

	entries, err := e.silenceStore.GetSilencesByName(ctx, event.Entity.Namespace, event.Check.Silenced)
	if err != nil {
		// Do not wrap this error, it needs to have its type inspected
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

	if err := e.silenceStore.DeleteSilences(ctx, event.Entity.Namespace, toDelete); err != nil {
		return fmt.Errorf("couldn't resolve silences: %s", err)
	}

	if len(event.Check.Silenced) != len(toRetain) {
		event.Check.Silenced = toRetain
		serialized, err := proto.Marshal(event)
		if err != nil {
			return &store.ErrEncode{Err: fmt.Errorf("couldn't resolve silences: %s", err)}
		}
		selectors := marshalSelectors(event)
		if _, err := e.db.Exec(ctx, UpdateEventOnlyQuery, serialized, selectors, event.Entity.Namespace, event.Entity.Name, event.Check.Name); err != nil {
			return &store.ErrInternal{Message: fmt.Sprintf("couldn't resolve silences: %s", err)}
		}
	}

	return nil
}

func (e *EventStore) CountEvents(ctx context.Context, pred *store.SelectionPredicate) (int64, error) {
	ns := corev2.ContextNamespace(ctx)
	if ns == corev2.NamespaceTypeAll {
		ns = ""
	}
	query, args, err := CreateCountEventsQuery(ns, selector.SelectorFromContext(ctx), pred)
	if err != nil {
		return 0, &store.ErrNotValid{Err: err}
	}
	row := e.db.QueryRow(ctx, query, args...)
	if err != nil {
		return 0, &store.ErrInternal{Message: fmt.Sprintf("couldn't get events: %s", err)}
	}
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (e *EventStore) EventStoreSupportsFiltering(_ context.Context) bool { return true }
