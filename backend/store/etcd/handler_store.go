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

func getHandlerPath(handler *types.Handler) string {
	return path.Join(etcdRoot, handlersPathPrefix, handler.Organization, handler.Name)
}

func getHandlersPath(ctx context.Context, name string) string {
	var org string

	// Determine the organization
	if value := ctx.Value(types.OrganizationKey); value != nil {
		org = value.(string)
	} else {
		org = ""
	}

	return path.Join(etcdRoot, handlersPathPrefix, org, name)
}

func (s *etcdStore) DeleteHandlerByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name of handler")
	}

	_, err := s.kvc.Delete(context.TODO(), getHandlersPath(ctx, name))
	return err
}

// GetHandlers gets the list of handlers for an (optional) organization. Passing
// the empty string as the org will return all handlers.
func (s *etcdStore) GetHandlers(ctx context.Context) ([]*types.Handler, error) {
	resp, err := s.kvc.Get(context.TODO(), getHandlersPath(ctx, ""), clientv3.WithPrefix())
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

func (s *etcdStore) GetHandlerByName(ctx context.Context, name string) (*types.Handler, error) {
	if name == "" {
		return nil, errors.New("must specify name of handler")
	}

	resp, err := s.kvc.Get(context.TODO(), getHandlersPath(ctx, name))
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

func (s *etcdStore) UpdateHandler(ctx context.Context, handler *types.Handler) error {
	if err := handler.Validate(); err != nil {
		return err
	}

	handlerBytes, err := json.Marshal(handler)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getHandlerPath(handler), string(handlerBytes))
	if err != nil {
		return err
	}

	return nil
}
