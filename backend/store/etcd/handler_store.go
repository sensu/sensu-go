package etcd

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

var (
	handlersPathPrefix = "handlers"
	handlerKeyBuilder  = store.NewKeyBuilder(handlersPathPrefix)
)

func getHandlerPath(handler *corev2.Handler) string {
	return handlerKeyBuilder.WithResource(handler).Build(handler.Name)
}

// GetHandlersPath gets the path of the handler store.
func GetHandlersPath(ctx context.Context, name string) string {
	return handlerKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteHandlerByName deletes a Handler by name.
func (s *Store) DeleteHandlerByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name of handler")}
	}

	err := Delete(ctx, s.client, GetHandlersPath(ctx, name))
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
	}
	return err
}

// GetHandlers gets the list of handlers for a namespace.
func (s *Store) GetHandlers(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Handler, error) {
	handlers := []*corev2.Handler{}
	err := List(ctx, s.client, GetHandlersPath, &handlers, pred)
	return handlers, err
}

// GetHandlerByName gets a Handler by name.
func (s *Store) GetHandlerByName(ctx context.Context, name string) (*corev2.Handler, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name of handler")}
	}

	var handler corev2.Handler
	if err := Get(ctx, s.client, GetHandlersPath(ctx, name), &handler); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
		return nil, err
	}
	return &handler, nil
}

// UpdateHandler updates a Handler.
func (s *Store) UpdateHandler(ctx context.Context, handler *corev2.Handler) error {
	if err := handler.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	return CreateOrUpdate(ctx, s.client, getHandlerPath(handler), handler.Namespace, handler)
}
