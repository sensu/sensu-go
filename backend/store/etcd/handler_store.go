package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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

// GetHandlersPath gets the path of the handler store.
func GetHandlersPath(ctx context.Context, name string) string {
	return handlerKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteHandlerByName deletes a Handler by name.
func (s *Store) DeleteHandlerByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name of handler")
	}

	_, err := s.client.Delete(ctx, GetHandlersPath(ctx, name))
	return err
}

// GetHandlers gets the list of handlers for a namespace.
func (s *Store) GetHandlers(ctx context.Context, pred *store.SelectionPredicate) ([]*types.Handler, error) {
	handlers := []*types.Handler{}
	err := List(ctx, s.client, GetHandlersPath, &handlers, pred)
	return handlers, err
}

// GetHandlerByName gets a Handler by name.
func (s *Store) GetHandlerByName(ctx context.Context, name string) (*types.Handler, error) {
	if name == "" {
		return nil, errors.New("must specify name of handler")
	}

	resp, err := s.client.Get(ctx, GetHandlersPath(ctx, name))
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
