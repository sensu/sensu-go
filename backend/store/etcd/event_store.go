package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	eventsPathPrefix = "events"
)

var (
	eventKeyBuilder = store.NewKeyBuilder(eventsPathPrefix)
)

func getEventPath(event *types.Event) string {
	return path.Join(
		EtcdRoot,
		eventsPathPrefix,
		event.Entity.Namespace,
		getEventKey(event),
	)
}

func getEventKey(event *types.Event) string {
	pathComponents := []string{}
	sortSkip := 0

	// Default to legacy behavior when Check.GroupBy is omitted
	if event.Check.GroupBy == nil {
		pathComponents = append(pathComponents, event.Entity.Name, event.Check.Name)
		sortSkip = 2
		goto join
	}

	if event.Check.GroupBy.Hostname {
		pathComponents = append(pathComponents, event.Entity.Name)
		sortSkip += 1
	}
	if event.Check.GroupBy.Name {
		pathComponents = append(pathComponents, event.Check.Name)
		sortSkip += 1
	}

	// The order of the elements in this struct are important. Changing them will break existing
	// store events.
	for _, group := range []struct {
		labels  map[string]string
		prefix  string
		include []string
	}{
		{
			labels:  event.Check.Labels,
			prefix:  "check",
			include: event.Check.GroupBy.CheckLabels,
		},
		{
			labels:  event.Entity.Labels,
			prefix:  "entity",
			include: event.Check.GroupBy.EntityLabels,
		},
		{
			labels:  event.Labels,
			prefix:  "event",
			include: event.Check.GroupBy.EventLabels,
		},
	} {
		for _, key := range group.include {
			pathComponents = append(pathComponents, fmt.Sprintf("%s:%s:%s", group.prefix, key, group.labels[key]))
		}
	}

join:
	// Exclude the name and hostname if they are present
	sort.Strings(pathComponents[sortSkip:])

	return path.Join(pathComponents...)
}

func getEventWithCheckPath(ctx context.Context, entity, check string) (string, error) {
	namespace := types.ContextNamespace(ctx)
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
func (s *Store) GetEvents(ctx context.Context) ([]*types.Event, error) {
	resp, err := s.client.Get(ctx, getEventsPath(ctx, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return []*types.Event{}, nil
	}

	var eventsArray []*types.Event
	for _, kv := range resp.Kvs {
		event := &types.Event{}
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

		eventsArray = append(eventsArray, event)
	}

	return eventsArray, nil
}

// GetEventsByEntity gets all events matching a given entity name.
func (s *Store) GetEventsByEntity(ctx context.Context, entityName string) ([]*types.Event, error) {
	if entityName == "" {
		return nil, errors.New("must specify entity name")
	}

	resp, err := s.client.Get(ctx, getEventsPath(ctx, entityName), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	eventsArray := make([]*types.Event, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		event := &types.Event{}
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
func (s *Store) GetEventByEntityCheck(ctx context.Context, entityName, checkName string) (*types.Event, error) {
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
	event := &types.Event{}
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
func (s *Store) UpdateEvent(ctx context.Context, event *types.Event) error {
	if event.Check == nil {
		return errors.New("event has no check")
	}

	if err := event.Check.Validate(); err != nil {
		return err
	}

	if err := event.Entity.Validate(); err != nil {
		return err
	}

	// Truncate check output if the output is larger than MaxOutputSize
	if size := event.Check.MaxOutputSize; size > 0 && int64(len(event.Check.Output)) > size {
		event.Check.Output = event.Check.Output[:size]
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
			"could not create the event %s in namespace %s",
			getEventKey(event),
			event.Entity.Namespace,
		)
	}

	return nil
}
