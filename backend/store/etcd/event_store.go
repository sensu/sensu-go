package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

const (
	eventsPathPrefix = "events"
)

func getEventsPath(org, entityID, checkID string) string {
	return path.Join(etcdRoot, eventsPathPrefix, org, entityID, checkID)
}

// Events

// GetEvents returns the events for an (optional) organization. If org is the
// empty string, GetEvents returns all events for all orgs.
func (s *etcdStore) GetEvents(org string) ([]*types.Event, error) {
	resp, err := s.kvc.Get(context.Background(), getEventsPath(org, "", ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return []*types.Event{}, nil
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

func (s *etcdStore) GetEventsByEntity(org, entityID string) ([]*types.Event, error) {
	if org == "" || entityID == "" {
		return nil, errors.New("must specify organization and entity id")
	}
	resp, err := s.kvc.Get(context.Background(), getEventsPath(org, entityID, ""), clientv3.WithPrefix())
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

func (s *etcdStore) GetEventByEntityCheck(org, entityID, checkID string) (*types.Event, error) {
	if org == "" || entityID == "" || checkID == "" {
		return nil, errors.New("must specify organization, entity, and check id")
	}

	resp, err := s.kvc.Get(context.Background(), getEventsPath(org, entityID, checkID), clientv3.WithPrefix())
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

func (s *etcdStore) UpdateEvent(event *types.Event) error {
	if event.Check == nil {
		return errors.New("event has no check")
	}

	// TODO(Simon): We should also validate event.Entity since we also use
	// some properties of Entity below, such as ID
	if err := event.Check.Validate(); err != nil {
		return err
	}

	// update the history
	// marshal the new event and store it.
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	org := event.Entity.Organization
	entityID := event.Entity.ID
	checkID := event.Check.Config.Name

	_, err = s.kvc.Put(context.TODO(), getEventsPath(org, entityID, checkID), string(eventBytes))
	if err != nil {
		return err
	}

	return nil
}

func (s *etcdStore) DeleteEventByEntityCheck(org, entityID, checkID string) error {
	if org == "" || entityID == "" || checkID == "" {
		return errors.New("must specify organization, entity, and check id")
	}

	_, err := s.kvc.Delete(context.TODO(), getEventsPath(org, entityID, checkID))
	return err
}
