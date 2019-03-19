package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
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

func getEventsPath(ctx context.Context, entity string) string {
	return eventKeyBuilder.WithContext(ctx).Build(entity)
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
func (s *Store) GetEvents(ctx context.Context, pageSize int64, continueToken string) (events []*corev2.Event, nextContinueToken string, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithLimit(pageSize),
	}

	keyPrefix := getEventsPath(ctx, "")
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	resp, err := s.client.Get(ctx, path.Join(keyPrefix, continueToken), opts...)
	if err != nil {
		return nil, "", err
	}

	if len(resp.Kvs) == 0 {
		return []*corev2.Event{}, "", nil
	}

	for _, kv := range resp.Kvs {
		event := &corev2.Event{}
		err = json.Unmarshal(kv.Value, event)
		if err != nil {
			return nil, "", err
		}
		if event.Labels == nil {
			event.Labels = make(map[string]string)
		}
		if event.Annotations == nil {
			event.Annotations = make(map[string]string)
		}

		events = append(events, event)
	}

	if resp.Count > pageSize {
		ns := store.NewNamespaceFromContext(ctx)
		lastEvent := events[len(events)-1]

		// TODO(ccressent): This can surely be simplified
		if ns == "" {
			nextContinueToken = "/" + lastEvent.Namespace + "/" + lastEvent.Entity.Name + "/" + lastEvent.Check.Name + "\x00"
		} else {
			nextContinueToken = lastEvent.Entity.Name + "/" + lastEvent.Check.Name + "\x00"
		}
	}

	return events, nextContinueToken, nil
}

// GetEventsByEntity gets all events matching a given entity name.
func (s *Store) GetEventsByEntity(ctx context.Context, entityName string) ([]*corev2.Event, error) {
	if entityName == "" {
		return nil, errors.New("must specify entity name")
	}

	opts := []clientv3.OpOption{
		clientv3.WithLimit(int64(corev2.PageSizeFromContext(ctx))),
	}

	keyPrefix := getEventsPath(ctx, entityName)
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	continueKey := corev2.PageContinueFromContext(ctx)

	resp, err := s.client.Get(ctx, path.Join(keyPrefix, continueKey), opts...)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	eventsArray := make([]*corev2.Event, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		event := &corev2.Event{}
		err = json.Unmarshal(kv.Value, event)
		if err != nil {
			return nil, err
		}
		if event.Labels == nil {
			event.Labels = make(map[string]string)
		}
		if event.Annotations == nil {
			event.Annotations = make(map[string]string)
		}
		eventsArray[i] = event
	}

	return eventsArray, nil
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
	if err := json.Unmarshal(eventBytes, event); err != nil {
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
func (s *Store) UpdateEvent(ctx context.Context, event *corev2.Event) error {
	if event == nil || event.Check == nil {
		return errors.New("event has no check")
	}

	if err := event.Check.Validate(); err != nil {
		return err
	}

	if err := event.Entity.Validate(); err != nil {
		return err
	}

	if event.HasMetrics() {
		// Taking pains to not modify our input, set metrics to nil so they are
		// not persisted.
		newEvent := *event
		event = &newEvent
		event.Metrics = nil
	}

	// Truncate check output if the output is larger than MaxOutputSize
	if size := event.Check.MaxOutputSize; size > 0 && int64(len(event.Check.Output)) > size {
		// Taking pains to not modify our input, set a bound on the check
		// output size.
		check := *event.Check
		check.Output = check.Output[:size]
		event.Check = &check
	}

	if event.Timestamp == 0 {
		// If the event is being created for the first time, it may not include
		// a timestamp. Use the current time.
		event.Timestamp = time.Now().Unix()
	}

	// update the history
	// marshal the new event and store it.
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	cmp := namespaceExistsForResource(event.Entity)
	req := clientv3.OpPut(getEventPath(event), string(eventBytes))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the event %s/%s in namespace %s",
			event.Entity.Name,
			event.Check.Name,
			event.Entity.Namespace,
		)
	}

	return nil
}
