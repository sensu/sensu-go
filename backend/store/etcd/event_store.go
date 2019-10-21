package etcd

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/sensu/sensu-go/backend/store"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	eventsPathPrefix = "events"
)

var (
	eventKeyBuilder = store.NewKeyBuilder(eventsPathPrefix)
)

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
		return errors.New("must specify entity and check name")
	}

	path, err := getEventWithCheckPath(ctx, entityName, checkName)
	if err != nil {
		return err
	}

	_, err = s.client.Delete(ctx, path)
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

	resp, err := s.client.Get(ctx, key, opts...)
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
			return nil, err
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
		return nil, errors.New("must specify entity name")
	}

	opts := []clientv3.OpOption{
		clientv3.WithLimit(pred.Limit),
	}

	keyPrefix := GetEventsPath(ctx, entityName)
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	resp, err := s.client.Get(ctx, path.Join(keyPrefix, pred.Continue), opts...)
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
			return nil, err
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
		return nil, errors.New("must specify entity and check name")
	}

	path, err := getEventWithCheckPath(ctx, entityName, checkName)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Get(ctx, path, clientv3.WithPrefix(), clientv3.WithSerializable())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	eventBytes := resp.Kvs[0].Value
	event := &corev2.Event{}
	if err := unmarshal(eventBytes, event); err != nil {
		return nil, err
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
		return nil, nil, errors.New("event has no check")
	}

	if err := event.Check.Validate(); err != nil {
		return nil, nil, err
	}

	if err := event.Entity.Validate(); err != nil {
		return nil, nil, err
	}

	ctx = store.NamespaceContext(ctx, event.Entity.Namespace)

	prevEvent, err := s.GetEventByEntityCheck(
		ctx, event.Entity.Name, event.Check.Name,
	)
	if err != nil {
		return nil, nil, err
	}

	// Maintain check history.
	if prevEvent != nil {
		if !prevEvent.HasCheck() {
			return nil, nil, errors.New("invalid previous event")
		}

		event.Check.MergeWith(prevEvent.Check)
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
	if err := handleExpireOnResolveEntries(ctx, persistEvent, s); err != nil {
		return nil, nil, err
	}

	// update the history
	// marshal the new event and store it.
	eventBytes, err := proto.Marshal(persistEvent)
	if err != nil {
		return nil, nil, err
	}

	cmp := namespaceExistsForResource(event.Entity)
	req := clientv3.OpPut(getEventPath(event), string(eventBytes))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return nil, nil, err
	}
	if !res.Succeeded {
		return nil, nil, fmt.Errorf(
			"could not create the event %s/%s in namespace %s",
			event.Entity.Name,
			event.Check.Name,
			event.Entity.Namespace,
		)
	}

	return event, prevEvent, nil
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

func handleExpireOnResolveEntries(ctx context.Context, event *corev2.Event, store store.Store) error {
	// Make sure we have a check and that the event is a resolution
	if !event.HasCheck() || !event.IsResolution() {
		return nil
	}

	entries, err := store.GetSilencedEntriesByName(ctx, event.Check.Silenced...)
	if err != nil {
		return fmt.Errorf("couldn't resolve silences: %s", err)
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

	if err := store.DeleteSilencedEntryByName(ctx, toDelete...); err != nil {
		return fmt.Errorf("couldn't resolve silences: %s", err)
	}
	event.Check.Silenced = toRetain

	return nil
}
