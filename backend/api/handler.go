package api

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// HandlerClient is an API client for handlers.
type HandlerClient struct {
	client GenericClient
	auth   authorization.Authorizer
}

// NewHandlerClient creates a new HandlerClient, given a store and authorizer.
func NewHandlerClient(store storev2.Interface, auth authorization.Authorizer) *HandlerClient {
	return &HandlerClient{
		client: GenericClient{
			Kind:       &corev2.Handler{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v2",
		},
		auth: auth,
	}
}

// ListHandlers fetches a list of handler resources, if authorized.
func (a *HandlerClient) ListHandlers(ctx context.Context) ([]*corev2.Handler, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.Handler{}
	if err := a.client.List(ctx, &slice, pred); err != nil {
		return nil, err
	}
	return slice, nil
}

// FetchHandler fetches a handler resource from the backend, if authorized.
func (a *HandlerClient) FetchHandler(ctx context.Context, name string) (*corev2.Handler, error) {
	var handler corev2.Handler
	if err := a.client.Get(ctx, name, &handler); err != nil {
		return nil, err
	}
	return &handler, nil
}

// CreateHandler creates a handler resource, if authorized.
func (a *HandlerClient) CreateHandler(ctx context.Context, handler *corev2.Handler) error {
	if err := a.client.Create(ctx, handler); err != nil {
		return err
	}
	return nil
}

// UpdateHandler updates a handler resource, if authorized.
func (a *HandlerClient) UpdateHandler(ctx context.Context, handler *corev2.Handler) error {
	if err := a.client.Update(ctx, handler); err != nil {
		return err
	}
	return nil
}

func (a *HandlerClient) DeleteHandler(ctx context.Context, name string) error {
	if err := a.client.Delete(ctx, name); err != nil {
		return err
	}
	return nil
}
