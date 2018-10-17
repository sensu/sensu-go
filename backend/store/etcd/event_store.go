package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

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
		event.Entity.ID,
		event.Check.Name,
	)
}

func getEventWithCheckPath(ctx context.Context, entity, check string) string {
	namespace := types.ContextNamespace(ctx)

	return path.Join(EtcdRoot, eventsPathPrefix, namespace, entity, check)
}

func getEventsPath(ctx context.Context, entity string) string {
	return eventKeyBuilder.WithContext(ctx).Build(entity)
}

// DeleteEventByEntityCheck deletes an event by entity ID and check ID.
func (s *Store) DeleteEventByEntityCheck(ctx context.Context, entityID, checkID string) error {
	if entityID == "" || checkID == "" {
		return errors.New("must specify entity and check id")
	}

	_, err := s.client.Delete(ctx, getEventWithCheckPath(ctx, entityID, checkID))
	return err
}

// GetEvents returns the events for an (optional) namespace. If namespace is the
// empty string, GetEvents returns all events for all namespaces.
func (s *Store) GetEvents(ctx context.Context) ([]*types.Event, error) {
	resp, err := query(ctx, s, getEventsPath)
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

		eventsArray = append(eventsArray, event)
	}

	return eventsArray, nil
}

// GetEventsByEntity gets all events matching a given entity ID.
func (s *Store) GetEventsByEntity(ctx context.Context, entityID string) ([]*types.Event, error) {
	if entityID == "" {
		return nil, errors.New("must specify entity id")
	}

	resp, err := s.client.Get(ctx, getEventsPath(ctx, entityID), clientv3.WithPrefix())
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
		eventsArray[i] = event
	}

	return eventsArray, nil
}

// GetEventByEntityCheck gets an event by entity and check ID.
func (s *Store) GetEventByEntityCheck(ctx context.Context, entityID, checkID string) (*types.Event, error) {
	if entityID == "" || checkID == "" {
		return nil, errors.New("must specify entity and check id")
	}

	resp, err := s.client.Get(ctx, getEventWithCheckPath(ctx, entityID, checkID), clientv3.WithPrefix())
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
			event.Entity.ID,
			event.Check.Name,
			event.Entity.Namespace,
		)
	}

	return nil
}
