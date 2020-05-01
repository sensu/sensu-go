package etcd

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

var (
	eventFiltersPathPrefix = "event-filters"
	eventFilterKeyBuilder  = store.NewKeyBuilder(eventFiltersPathPrefix)
)

func getEventFilterPath(filter *corev2.EventFilter) string {
	return eventFilterKeyBuilder.WithResource(filter).Build(filter.Name)
}

// GetEventFiltersPath gets the path of the event filter store.
func GetEventFiltersPath(ctx context.Context, name string) string {
	return eventFilterKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteEventFilterByName deletes an EventFilter by name.
func (s *Store) DeleteEventFilterByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name of filter")}
	}

	err := Delete(ctx, s.client, GetEventFiltersPath(ctx, name))
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
	}
	return err
}

// GetEventFilters gets the list of filters for a namespace.
func (s *Store) GetEventFilters(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.EventFilter, error) {
	filters := []*corev2.EventFilter{}
	err := List(ctx, s.client, GetEventFiltersPath, &filters, pred)
	return filters, err
}

// GetEventFilterByName gets an EventFilter by name.
func (s *Store) GetEventFilterByName(ctx context.Context, name string) (*corev2.EventFilter, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name of filter")}
	}

	var filter corev2.EventFilter
	if err := Get(ctx, s.client, GetEventFiltersPath(ctx, name), &filter); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
		return nil, err
	}

	return &filter, nil
}

// UpdateEventFilter updates an EventFilter.
func (s *Store) UpdateEventFilter(ctx context.Context, filter *corev2.EventFilter) error {
	if err := filter.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	return CreateOrUpdate(ctx, s.client, getEventFilterPath(filter), filter.Namespace, filter)
}
