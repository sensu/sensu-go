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
	handlersPathPrefix = "handlers"
)

func getHandlersPath(org, name string) string {
	return path.Join(etcdRoot, handlersPathPrefix, org, name)
}

// Handlers gets the list of handlers for an (optional) organization. Passing
// the empty string as the org will return all handlers.
func (s *etcdStore) GetHandlers(org string) ([]*types.Handler, error) {
	resp, err := s.kvc.Get(context.TODO(), getHandlersPath(org, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*types.Handler{}, nil
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

func (s *etcdStore) GetHandlerByName(org, name string) (*types.Handler, error) {
	if org == "" || name == "" {
		return nil, errors.New("must specify organization and name of handler")
	}
	resp, err := s.kvc.Get(context.TODO(), getHandlersPath(org, name))
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

func (s *etcdStore) DeleteHandlerByName(org, name string) error {
	if org == "" || name == "" {
		return errors.New("must specify organization and name of handler")
	}
	_, err := s.kvc.Delete(context.TODO(), getHandlersPath(org, name))
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

	_, err = s.kvc.Put(context.TODO(), getHandlersPath(handler.Organization, handler.Name), string(handlerBytes))
	if err != nil {
		return err
	}

	return nil
}
