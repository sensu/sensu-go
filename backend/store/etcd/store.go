package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	etcdRoot = "/sensu.io"
)

var logger *logrus.Entry

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"component": "store",
	})
}

func getEntityPath(id string) string {
	return fmt.Sprintf("%s/entities/%s", etcdRoot, id)
}

func getHandlersPath(name string) string {
	return fmt.Sprintf("%s/handlers/%s", etcdRoot, name)
}

func getMutatorsPath(name string) string {
	return fmt.Sprintf("%s/mutators/%s", etcdRoot, name)
}

func getChecksPath(name string) string {
	return fmt.Sprintf("%s/checks/%s", etcdRoot, name)
}

func getEventsPath(args ...string) string {
	return fmt.Sprintf("%s/events/%s", etcdRoot, strings.Join(args, "/"))
}

// Store is an implementation of the sensu-go/backend/store.Store iface.
type etcdStore struct {
	client *clientv3.Client
	kvc    clientv3.KV
	etcd   *Etcd
}

func (s *etcdStore) UpdateEntity(e *types.Entity) error {
	eStr, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.kvc.Put(context.TODO(), getEntityPath(e.ID), string(eStr))
	return err
}

func (s *etcdStore) DeleteEntity(e *types.Entity) error {
	_, err := s.kvc.Delete(context.TODO(), getEntityPath(e.ID))
	return err
}

func (s *etcdStore) GetEntityByID(id string) (*types.Entity, error) {
	resp, err := s.kvc.Get(context.TODO(), getEntityPath(id), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, nil
	}
	entity := &types.Entity{}
	err = json.Unmarshal(resp.Kvs[0].Value, entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *etcdStore) GetEntities() ([]*types.Entity, error) {
	resp, err := s.kvc.Get(context.TODO(), getEntityPath(""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	earr := make([]*types.Entity, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		entity := &types.Entity{}
		err = json.Unmarshal(kv.Value, entity)
		if err != nil {
			return nil, err
		}
		earr[i] = entity
	}

	return earr, nil
}

// Handlers
func (s *etcdStore) GetHandlers() ([]*types.Handler, error) {
	resp, err := s.kvc.Get(context.TODO(), getHandlersPath(""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	handlersArray := make([]*types.Handler, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		handler := &types.Handler{}
		err = json.Unmarshal(kv.Value, handler)
		if err != nil {
			return nil, err
		}
		handlersArray[i] = handler
	}

	return handlersArray, nil
}

func (s *etcdStore) GetHandlerByName(name string) (*types.Handler, error) {
	resp, err := s.kvc.Get(context.TODO(), getHandlersPath(name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	handlerBytes := resp.Kvs[0].Value
	handler := &types.Handler{}
	if err := json.Unmarshal(handlerBytes, handler); err != nil {
		return nil, err
	}

	return handler, nil
}

func (s *etcdStore) DeleteHandlerByName(name string) error {
	_, err := s.kvc.Delete(context.TODO(), getHandlersPath(name))
	return err
}

func (s *etcdStore) UpdateHandler(handler *types.Handler) error {
	if err := handler.Validate(); err != nil {
		return err
	}

	handlerBytes, err := json.Marshal(handler)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getHandlersPath(handler.Name), string(handlerBytes))
	if err != nil {
		return err
	}

	return nil
}

// Mutators
func (s *etcdStore) GetMutators() ([]*types.Mutator, error) {
	resp, err := s.kvc.Get(context.TODO(), getMutatorsPath(""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	mutatorsArray := make([]*types.Mutator, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		mutator := &types.Mutator{}
		err = json.Unmarshal(kv.Value, mutator)
		if err != nil {
			return nil, err
		}
		mutatorsArray[i] = mutator
	}

	return mutatorsArray, nil
}

func (s *etcdStore) GetMutatorByName(name string) (*types.Mutator, error) {
	resp, err := s.kvc.Get(context.TODO(), getMutatorsPath(name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	mutatorBytes := resp.Kvs[0].Value
	mutator := &types.Mutator{}
	if err := json.Unmarshal(mutatorBytes, mutator); err != nil {
		return nil, err
	}

	return mutator, nil
}

func (s *etcdStore) DeleteMutatorByName(name string) error {
	_, err := s.kvc.Delete(context.TODO(), getMutatorsPath(name))
	return err
}

func (s *etcdStore) UpdateMutator(mutator *types.Mutator) error {
	if err := mutator.Validate(); err != nil {
		return err
	}

	mutatorBytes, err := json.Marshal(mutator)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getMutatorsPath(mutator.Name), string(mutatorBytes))
	if err != nil {
		return err
	}

	return nil
}

// Checks
func (s *etcdStore) GetChecks() ([]*types.Check, error) {
	resp, err := s.kvc.Get(context.TODO(), getChecksPath(""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	checksArray := make([]*types.Check, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		check := &types.Check{}
		err = json.Unmarshal(kv.Value, check)
		if err != nil {
			return nil, err
		}
		checksArray[i] = check
	}

	return checksArray, nil
}

func (s *etcdStore) GetCheckByName(name string) (*types.Check, error) {
	resp, err := s.kvc.Get(context.TODO(), getChecksPath(name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	checkBytes := resp.Kvs[0].Value
	check := &types.Check{}
	if err := json.Unmarshal(checkBytes, check); err != nil {
		return nil, err
	}

	return check, nil
}

func (s *etcdStore) DeleteCheckByName(name string) error {
	_, err := s.kvc.Delete(context.TODO(), getChecksPath(name))
	return err
}

func (s *etcdStore) UpdateCheck(check *types.Check) error {
	if err := check.Validate(); err != nil {
		return err
	}

	checkBytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getChecksPath(check.Name), string(checkBytes))
	if err != nil {
		return err
	}

	return nil
}

// Events

func (s *etcdStore) GetEvents() ([]*types.Event, error) {
	resp, err := s.kvc.Get(context.Background(), getEventsPath(""), clientv3.WithPrefix())
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

func (s *etcdStore) GetEventsByEntity(entityID string) ([]*types.Event, error) {
	resp, err := s.kvc.Get(context.Background(), getEventsPath(entityID), clientv3.WithPrefix())
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

func (s *etcdStore) GetEventByEntityCheck(entityID, checkID string) (*types.Event, error) {
	if entityID == "" {
		return nil, errors.New("entity id is required")
	}

	if checkID == "" {
		return nil, errors.New("check id is required")
	}

	resp, err := s.kvc.Get(context.Background(), getEventsPath(entityID, checkID), clientv3.WithPrefix())
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

	if err := event.Check.Validate(); err != nil {
		return err
	}

	// update the history
	// marshal the new event and store it.
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	entityID := event.Entity.ID
	checkID := event.Check.Name

	_, err = s.kvc.Put(context.TODO(), getEventsPath(entityID, checkID), string(eventBytes))
	if err != nil {
		return err
	}

	return nil
}

func (s *etcdStore) DeleteEventByEntityCheck(entityID, checkID string) error {
	if entityID == "" {
		return errors.New("entity id is required")
	}

	if checkID == "" {
		return errors.New("check id is required")
	}

	_, err := s.kvc.Delete(context.TODO(), getEventsPath(entityID, checkID))
	return err
}

// NewStore ...
func (e *Etcd) NewStore() (store.Store, error) {
	c, err := e.NewClient()
	if err != nil {
		return nil, err
	}

	store := &etcdStore{
		etcd:   e,
		client: c,
		kvc:    clientv3.NewKV(c),
	}
	return store, nil
}
