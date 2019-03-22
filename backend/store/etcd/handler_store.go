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

var (
	handlersPathPrefix = "handlers"
	handlerKeyBuilder  = store.NewKeyBuilder(handlersPathPrefix)
)

func getHandlerPath(handler *types.Handler) string {
	return handlerKeyBuilder.WithResource(handler).Build(handler.Name)
}

func getHandlersPath(ctx context.Context, name string) string {
	return handlerKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteHandlerByName deletes a Handler by name.
func (s *Store) DeleteHandlerByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name of handler")
	}

	_, err := s.client.Delete(ctx, getHandlersPath(ctx, name))
	return err
}

// GetHandlers gets the list of handlers for a namespace.
func (s *Store) GetHandlers(ctx context.Context, pageSize int64, continueToken string) (handlers []*types.Handler, newContinueToken string, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithLimit(pageSize),
	}

	keyPrefix := getHandlersPath(ctx, "")
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	resp, err := s.client.Get(ctx, path.Join(keyPrefix, continueToken), opts...)
	if err != nil {
		return nil, "", err
	}
	if len(resp.Kvs) == 0 {
		return []*types.Handler{}, "", nil
	}

	for _, kv := range resp.Kvs {
		handler := &types.Handler{}
		err = json.Unmarshal(kv.Value, handler)
		if err != nil {
			return nil, "", err
		}

		handlers = append(handlers, handler)
	}

	if pageSize != 0 && resp.Count > pageSize {
		lastHandler := handlers[len(handlers)-1]
		newContinueToken = lastHandler.Name + "\x00"
	}

	return handlers, newContinueToken, nil
}

// GetHandlerByName gets a Handler by name.
func (s *Store) GetHandlerByName(ctx context.Context, name string) (*types.Handler, error) {
	if name == "" {
		return nil, errors.New("must specify name of handler")
	}

	resp, err := s.client.Get(ctx, getHandlersPath(ctx, name))
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

// UpdateHandler updates a Handler.
func (s *Store) UpdateHandler(ctx context.Context, handler *types.Handler) error {
	if err := handler.Validate(); err != nil {
		return err
	}

	handlerBytes, err := json.Marshal(handler)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(handler.Namespace)), ">", 0)
	req := clientv3.OpPut(getHandlerPath(handler), string(handlerBytes))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the handler %s in namespace %s",
			handler.Name,
			handler.Namespace,
		)
	}

	return nil
}
