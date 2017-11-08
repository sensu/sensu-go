package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

const (
	eventsPathPrefix = "events"
)

func getEventPath(event *types.Event) string {
	return path.Join(
		etcdRoot,
		eventsPathPrefix,
		event.Entity.Organization,
		event.Entity.Environment,
		event.Entity.ID,
		event.Check.Config.Name,
	)
}

func getEventsPath(ctx context.Context, entity, check string) string {
	env := environment(ctx)
	org := organization(ctx)

	return path.Join(etcdRoot, eventsPathPrefix, org, env, entity, check)
}

func (s *etcdStore) DeleteEventByEntityCheck(ctx context.Context, entityID, checkID string) error {
	if entityID == "" || checkID == "" {
		return errors.New("must specify entity and check id")
	}

	_, err := s.kvc.Delete(ctx, getEventsPath(ctx, entityID, checkID))
	return err
}

// GetEvents returns the events for an (optional) organization. If org is the
// empty string, GetEvents returns all events for all orgs.
func (s *etcdStore) GetEvents(ctx context.Context) ([]*types.Event, error) {
	// TODO (SP): We should use the query function here but getEnvironmentsPath signature is wrong
	resp, err := s.kvc.Get(context.Background(), getEventsPath(ctx, "", ""), clientv3.WithPrefix())
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

func (s *etcdStore) GetEventsByEntity(ctx context.Context, entityID string) ([]*types.Event, error) {
	if entityID == "" {
		return nil, errors.New("must specify entity id")
	}

	resp, err := s.kvc.Get(context.Background(), getEventsPath(ctx, entityID, ""), clientv3.WithPrefix())
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

func (s *etcdStore) GetEventByEntityCheck(ctx context.Context, entityID, checkID string) (*types.Event, error) {
	if entityID == "" || checkID == "" {
		return nil, errors.New("must specify entity and check id")
	}

	resp, err := s.kvc.Get(context.Background(), getEventsPath(ctx, entityID, checkID), clientv3.WithPrefix())
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

func (s *etcdStore) UpdateEvent(ctx context.Context, event *types.Event) error {
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

	cmp := clientv3.Compare(clientv3.Version(getEnvironmentsPath(event.Entity.Organization, event.Entity.Environment)), ">", 0)
	req := clientv3.OpPut(getEventPath(event), string(eventBytes))
	res, err := s.kvc.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the event %s/%s in environment %s/%s",
			event.Entity.ID,
			event.Check.Config.Name,
			event.Entity.Organization,
			event.Entity.Environment,
		)
	}

	return nil
}
