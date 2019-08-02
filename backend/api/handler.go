package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
)

type HandlerClient struct {
	client genericClient
	auth   authorization.Authorizer
}

func NewHandlerClient(store store.ResourceStore, auth authorization.Authorizer) *HandlerClient {
	return &HandlerClient{
		client: genericClient{
			Kind:       &corev2.Handler{},
			Store:      store,
			Auth:       auth,
			Resource:   "handlers",
			APIGroup:   "core",
			APIVersion: "v2",
		},
		auth: auth,
	}
}

// ListHandlers fetches a list of handler resources
func (a *HandlerClient) ListHandlers(ctx context.Context) ([]*corev2.Handler, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.Handler{}
	if err := a.client.List(ctx, &slice, pred); err != nil {
		return nil, fmt.Errorf("couldn't list handlers: %s", err)
	}
	return slice, nil
}

// FetchHandler fetches a handler resource from the backend
func (a *HandlerClient) FetchHandler(ctx context.Context, name string) (*corev2.Handler, error) {
	var handler corev2.Handler
	if err := a.client.Get(ctx, name, &handler); err != nil {
		return nil, fmt.Errorf("couldn't get handler: %s", err)
	}
	return &handler, nil
}

// CreateHandler creates a handler resource
func (a *HandlerClient) CreateHandler(ctx context.Context, handler *corev2.Handler) error {
	if err := a.client.Create(ctx, handler); err != nil {
		return fmt.Errorf("couldn't create handler: %s", err)
	}
	return nil
}

// UpdateHandler updates a handler resource
func (a *HandlerClient) UpdateHandler(ctx context.Context, handler *corev2.Handler) error {
	if err := a.client.Update(ctx, handler); err != nil {
		return fmt.Errorf("couldn't update handler: %s", err)
	}
	return nil
}
