package etcd

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sensu/sensu-go/backend/metrics"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/kvc"
	"github.com/sensu/sensu-go/backend/store/provider"
	clientv3 "go.etcd.io/etcd/client/v3"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	eventsPathPrefix = "events"
	// Type is the type of an etcd store provider.
	Type = "etcd"

	// EventBytesSummaryName is the name of the prometheus summary vec used to
	// track event sizes (in bytes).
	EventBytesSummaryName = "sensu_go_store_event_bytes"

	// EventBytesSummaryHelp is the help message for EventBytesSummary
	// Prometheus metrics.
	EventBytesSummaryHelp = "Distribution of event sizes, in bytes, received by the store on this backend"

	eventOutputTruncatedBytesLabel = "sensu.io/output_truncated_bytes"
)

var (
	eventKeyBuilder = store.NewKeyBuilder(eventsPathPrefix)

	EventBytesSummary = metrics.NewEventBytesSummaryVec(EventBytesSummaryName, EventBytesSummaryHelp)
)

func init() {
	if err := prometheus.Register(EventBytesSummary); err != nil {
		metrics.LogError(logger, EventBytesSummaryName, err)
	}
}

func getEventPath(event *corev2.Event) string {
	return path.Join(
		EtcdRoot,
		eventsPathPrefix,
		event.Entity.Namespace,
		event.Entity.Name,
		event.Check.Name,
	)
}

func getEventWithCheckPath(ctx context.Context, entity, check string) (string, error) {
	namespace := corev2.ContextNamespace(ctx)
	if namespace == "" {
		return "", errors.New("namespace missing from context")
	}

	return path.Join(EtcdRoot, eventsPathPrefix, namespace, entity, check), nil
}

// GetEventsPath gets the path of the event store.
func GetEventsPath(ctx context.Context, entity string) string {
	b := eventKeyBuilder.WithContext(ctx)
	if entity != "" {
		b = b.WithExactMatch()
	}
	return b.Build(entity)
}

// DeleteEventByEntityCheck deletes an event by entity name and check name.
func (s *Store) DeleteEventByEntityCheck(ctx context.Context, entityName, checkName string) error {
	if entityName == "" || checkName == "" {
		return &store.ErrNotValid{Err: errors.New("must specify entity and check name")}
	}

	path, err := getEventWithCheckPath(ctx, entityName, checkName)
	if err != nil {
		return &store.ErrNotValid{Err: err}
	}

	err = Delete(ctx, s.client, path)
	if _, ok := err.(*store.ErrNotFound); ok {
		err = nil
	}
	return err
}

// GetEvents returns the events for an (optional) namespace. If namespace is the
// empty string, GetEvents returns all events for all namespaces.
func (s *Store) GetEvents(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	opts := []clientv3.OpOption{
		clientv3.WithLimit(pred.Limit),
	}

	keyPrefix := GetEventsPath(ctx, "")
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	key := keyPrefix
	if pred.Continue != "" {
		key = path.Join(keyPrefix, pred.Continue)
	} else {
		if !strings.HasSuffix(key, "/") {
			key += "/"
		}
	}

	var resp *clientv3.GetResponse
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, key, opts...)
		return kvc.RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return []*corev2.Event{}, nil
	}

	events := []*corev2.Event{}
	for _, kv := range resp.Kvs {
		event := &corev2.Event{}
		if err := unmarshal(kv.Value, event); err != nil {
			return nil, &store.ErrDecode{Err: err}
		}

		if event.Labels == nil {
			event.Labels = make(map[string]string)
		}
		if event.Annotations == nil {
			event.Annotations = make(map[string]string)
		}

		events = append(events, event)
	}

	if pred.Limit != 0 && resp.Count > pred.Limit {
		pred.Continue = ComputeContinueToken(ctx, events[len(events)-1])
	} else {
		pred.Continue = ""
	}

	return events, nil
}

// GetEventsByEntity gets all events matching a given entity name.
func (s *Store) GetEventsByEntity(ctx context.Context, entityName string, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	if entityName == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify entity name")}
	}

	opts := []clientv3.OpOption{
		clientv3.WithLimit(pred.Limit),
	}

	keyPrefix := GetEventsPath(ctx, entityName)
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	var resp *clientv3.GetResponse
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, fmt.Sprintf("%s/", path.Join(keyPrefix, pred.Continue)), opts...)
		return kvc.RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	events := []*corev2.Event{}
	for _, kv := range resp.Kvs {
		event := &corev2.Event{}
		if err := unmarshal(kv.Value, event); err != nil {
			return nil, &store.ErrDecode{Err: err}
		}

		if event.Labels == nil {
			event.Labels = make(map[string]string)
		}
		if event.Annotations == nil {
			event.Annotations = make(map[string]string)
		}

		events = append(events, event)
	}

	if pred.Limit != 0 && resp.Count > pred.Limit {
		lastEvent := events[len(events)-1]
		pred.Continue = lastEvent.Check.Name + "\x00"
	} else {
		pred.Continue = ""
	}

	return events, nil
}

// GetEventByEntityCheck gets an event by entity and check name.
func (s *Store) GetEventByEntityCheck(ctx context.Context, entityName, checkName string) (*corev2.Event, error) {
	if entityName == "" || checkName == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify entity and check name")}
	}

	path, err := getEventWithCheckPath(ctx, entityName, checkName)
	if err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}

	var resp *clientv3.GetResponse
	err = kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, path, clientv3.WithPrefix(), clientv3.WithSerializable())
		return kvc.RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	eventBytes := resp.Kvs[0].Value
	event := &corev2.Event{}
	if err := unmarshal(eventBytes, event); err != nil {
		return nil, &store.ErrDecode{Err: err}
	}

	if event.Labels == nil {
		event.Labels = make(map[string]string)
	}
	if event.Annotations == nil {
		event.Annotations = make(map[string]string)
	}

	return event, nil
}

// UpdateEvent updates an event.
func (s *Store) UpdateEvent(ctx context.Context, event *corev2.Event) (*corev2.Event, *corev2.Event, error) {
	if event == nil || event.Check == nil {
		return nil, nil, &store.ErrNotValid{Err: errors.New("event has no check")}
	}

	if err := event.Check.Validate(); err != nil {
		return nil, nil, &store.ErrNotValid{Err: err}
	}

	if err := event.Entity.Validate(); err != nil {
		return nil, nil, &store.ErrNotValid{Err: err}
	}

	ctx = store.NamespaceContext(ctx, event.Entity.Namespace)

	prevEvent, err := s.GetEventByEntityCheck(
		ctx, event.Entity.Name, event.Check.Name,
	)
	if err != nil {
		return nil, nil, err
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
	if err := handleExpireOnResolveEntries(ctx, persistEvent, s); err != nil {
		return nil, nil, err
	}

	// update the history
	// marshal the new event and store it.
	eventBytes, err := proto.Marshal(persistEvent)
	if err != nil {
		return nil, nil, &store.ErrEncode{Err: err}
	}

	EventBytesSummary.WithLabelValues(typeLabelValue).Observe(float64(len(eventBytes)))

	cmp := namespaceExistsForResource(event.Entity)
	req := clientv3.OpPut(getEventPath(event), string(eventBytes))
	var res *clientv3.TxnResponse
	err = kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		res, err = s.client.Txn(ctx).If(cmp).Then(req).Commit()
		return kvc.RetryRequest(n, err)
	})
	if err != nil {
		return nil, nil, err
	}
	if !res.Succeeded {
		return nil, nil, &store.ErrNamespaceMissing{Namespace: event.Entity.Namespace}
	}

	return event, prevEvent, nil
}

// GetProviderInfo returns the info of an etcd store provider.
func (s *Store) GetProviderInfo() *provider.Info {
	return &provider.Info{
		TypeMeta: corev2.TypeMeta{
			Type:       Type,
			APIVersion: "store/v1",
		},
		ObjectMeta: corev2.ObjectMeta{
			Name: Type,
		},
	}
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

// CountEvents counts events in the namespace. The SelectionPredicate is not
// supported.
func (s *Store) CountEvents(ctx context.Context, _ *store.SelectionPredicate) (int64, error) {
	key := GetEventsPath(ctx, "")
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}

	return Count(ctx, s.client, key)
}

// EventStoreSupportsFiltering the etcd store does not currently support
// filtering.
func (s *Store) EventStoreSupportsFiltering(context.Context) bool {
	return false
}
