package etcd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/metrics"
	"github.com/sensu/sensu-go/backend/store"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	eventsPathPrefix = "events"
)

var (
	eventKeyBuilder = store.NewKeyBuilder(eventsPathPrefix)
	txnFailedError  = errors.New("transaction failed")
	eventBatchSize  = 5
)

func init() {
	bs, err := strconv.Atoi(os.Getenv("SENSU_BATCH_SIZE"))
	if err == nil {
		eventBatchSize = bs
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

func getEventOp(ctx context.Context, event *corev2.Event) (clientv3.Op, error) {
	var op clientv3.Op
	if event == nil || event.Check == nil || event.Entity == nil {
		return op, errors.New("invalid event")
	}
	if event.Entity.Name == "" || event.Check.Name == "" {
		return op, errors.New("must specify entity and check name")
	}

	path, err := getEventWithCheckPath(ctx, event.Entity.Name, event.Check.Name)
	if err != nil {
		return op, err
	}

	return clientv3.OpGet(path, clientv3.WithPrefix()), nil
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

	resp, err := s.client.Get(ctx, path, clientv3.WithPrefix())
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

func (s *Store) UpdateEventBatch(ctx context.Context, event *corev2.Event) error {
	s.eventBatcherOnce.Do(func() {
		for i := 0; i < 100; i++ {
			go s.startEventBatcher()
		}
	})
	if err := s.eventQueue.Send(ctx, event); err != nil {
		return err
	}
	return nil
}

func (s *Store) startEventBatcher() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	batch := make([]*corev2.Event, 0, eventBatchSize)
	for {
		select {
		case <-ticker.C:
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			if len(batch) > 0 {
				s.handleEventBatch(context.TODO(), batch)
				metrics.EventsProcessed.WithLabelValues(metrics.EventsProcessedLabelSuccess).Add(float64(len(batch)))
				batch = batch[0:0]
			}
		default:
			event, err := s.eventQueue.Receive(ctx)
			if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
				logger.WithError(err).Error("error while writing batched event")
				continue
			} else if err == nil {
				batch = append(batch, event)
				if len(batch) >= eventBatchSize {
					s.handleEventBatch(context.TODO(), batch)
					metrics.EventsProcessed.WithLabelValues(metrics.EventsProcessedLabelSuccess).Add(float64(len(batch)))
					batch = batch[0:0]
				}
			}
		}
	}
}

func (s *Store) handleEventBatch(ctx context.Context, inputEvents []*corev2.Event) {
	currentEvents, err := s.getCurrentEventsBatch(ctx, inputEvents)
	if err != nil {
		logger.WithError(err).Error("couldn't write events")
		return
	}
	cmps, ops, err := s.getBatchOps(ctx, inputEvents, currentEvents)
	if err != nil {
		logger.WithError(err).Error("couldn't write events")
		return
	}
	resp, err := s.client.Txn(ctx).If(cmps...).Then(ops...).Commit()
	if err != nil {
		logger.WithError(err).Error("couldn't write events")
		return
	}
	if !resp.Succeeded {
		logger.WithError(err).Error("couldn't write events")
		return
	}
}

func (s *Store) getBatchOps(ctx context.Context, inputEvents, currentEvents []*corev2.Event) ([]clientv3.Cmp, []clientv3.Op, error) {
	if len(inputEvents) != len(currentEvents) {
		return nil, nil, errors.New("mismatched input and current events")
	}
	cmps := make([]clientv3.Cmp, 0, len(inputEvents))
	ops := make([]clientv3.Op, 0, len(inputEvents))
	for i := range inputEvents {
		cmp, op, err := s.getEventPutOp(ctx, inputEvents[i], currentEvents[i])
		if err != nil {
			logger.WithError(err).Error("error creating op, skipping event")
			continue
		}
		cmps = append(cmps, cmp)
		ops = append(ops, op)
	}
	return cmps, ops, nil
}

// eventKey creates a key to identify the event for liveness monitoring
func eventKey(event *corev2.Event) string {
	// Typically we want the entity name to be the thing we monitor, but if
	// it's a round robin check, and there is no proxy entity, then use
	// the check name instead.
	if event.Check.RoundRobin && event.Entity.EntityClass != corev2.EntityProxyClass {
		return path.Join(event.Check.Namespace, event.Check.Name)
	}
	return path.Join(event.Entity.Namespace, event.Check.Name, event.Entity.Name)
}

func (s *Store) getEventPutOp(ctx context.Context, event, prevEvent *corev2.Event) (clientv3.Cmp, clientv3.Op, error) {
	var cmp clientv3.Cmp
	var op clientv3.Op
	if event == nil || event.Check == nil {
		return cmp, op, errors.New("event has no check")
	}

	if err := event.Check.Validate(); err != nil {
		return cmp, op, err
	}

	if err := event.Entity.Validate(); err != nil {
		return cmp, op, err
	}

	ctx = store.NamespaceContext(ctx, event.Entity.Namespace)

	// Maintain check history.
	if prevEvent != nil {
		if !prevEvent.HasCheck() {
			return cmp, op, errors.New("invalid previous event")
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

	// update the history
	// marshal the new event and store it.
	eventBytes, err := proto.Marshal(persistEvent)
	if err != nil {
		return cmp, op, err
	}

	cmp = namespaceExistsForResource(event.Entity)
	op = clientv3.OpPut(getEventPath(event), string(eventBytes))

	switches := liveness.WaitLookup(context.Background(), s.client, "eventd")
	switchKey := eventKey(event)

	if event.Check.Ttl > 0 {
		// Reset the switch
		timeout := int64(event.Check.Ttl)
		if err := switches.Alive(context.TODO(), switchKey, timeout); err != nil {
			logger.WithError(err).Error("error refreshing check TTL")
		}
	} else if prevEvent != nil && prevEvent.Check.Ttl > 0 {
		// The check TTL has been disabled, there is no longer a need to track it
		if err := switches.Bury(context.TODO(), switchKey); err != nil {
			logger.WithError(err).Error("error cancelling check TTL")
		}
	}

	return cmp, op, nil
}

func (s *Store) getCurrentEventsBatch(ctx context.Context, inputEvents []*corev2.Event) ([]*corev2.Event, error) {
	ops := make([]clientv3.Op, 0, len(inputEvents))
	for _, event := range inputEvents {
		ctx := store.NamespaceContext(ctx, event.Entity.Namespace)
		op, err := getEventOp(ctx, event)
		if err != nil {
			return nil, fmt.Errorf("error getting current events: %s", err)
		}
		ops = append(ops, op)
	}
	response, err := s.client.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return nil, fmt.Errorf("error reading event batch: %s", err)
	}
	if !response.Succeeded {
		return nil, txnFailedError
	}
	results := make([]*corev2.Event, 0, len(inputEvents))
	for _, r := range response.Responses {
		resp := r.GetResponseRange()
		if len(resp.Kvs) == 0 {
			results = append(results, nil)
			continue
		}
		for _, kv := range resp.Kvs {
			var event corev2.Event
			if err := unmarshal(kv.Value, &event); err != nil {
				return nil, fmt.Errorf("corrupted event: %s", err)
			}
			results = append(results, &event)
		}
	}
	return results, nil
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
