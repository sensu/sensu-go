package api

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// EventFilterClient is an API client for event filters.
type EventFilterClient struct {
	client GenericClient
	auth   authorization.Authorizer
}

// NewEventFilterClient creates a new EventFilterClient, given a store and
// authorizer.
func NewEventFilterClient(store storev2.Interface, auth authorization.Authorizer) *EventFilterClient {
	return &EventFilterClient{
		client: GenericClient{
			Store:      store,
			Auth:       auth,
			Kind:       &corev2.EventFilter{},
			APIGroup:   "core",
			APIVersion: "v2",
		},
		auth: auth,
	}
}

// ListEventFilters fetches a list of filter resources, if authorized.
func (a *EventFilterClient) ListEventFilters(ctx context.Context) ([]*corev2.EventFilter, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.EventFilter{}
	if err := a.client.List(ctx, &slice, pred); err != nil {
		return nil, err
	}
	return slice, nil
}

// FetchEventFilter fetches a filter resource from the backend, if authorized.
func (a *EventFilterClient) FetchEventFilter(ctx context.Context, name string) (*corev2.EventFilter, error) {
	var filter corev2.EventFilter
	if err := a.client.Get(ctx, name, &filter); err != nil {
		return nil, err
	}
	return &filter, nil
}

// CreateEventFilter creates a filter resource, if authorized.
func (a *EventFilterClient) CreateEventFilter(ctx context.Context, filter *corev2.EventFilter) error {
	if err := a.client.Create(ctx, filter); err != nil {
		return err
	}
	return nil
}

// UpdateEventFilter updates a filter resource, if authorized.
func (a *EventFilterClient) UpdateEventFilter(ctx context.Context, filter *corev2.EventFilter) error {
	if err := a.client.Update(ctx, filter); err != nil {
		return err
	}
	return nil
}

// DeleteEventFilter deletes a filter resource, if authorized.
func (a *EventFilterClient) DeleteEventFilter(ctx context.Context, name string) error {
	if err := a.client.Delete(ctx, name); err != nil {
		return err
	}
	return nil
}
