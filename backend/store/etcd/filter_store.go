package etcd

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	eventFiltersPathPrefix = "event-filters"
	eventFilterKeyBuilder  = store.NewKeyBuilder(eventFiltersPathPrefix)
)

func getEventFilterPath(filter *types.EventFilter) string {
	return eventFilterKeyBuilder.WithResource(filter).Build(filter.Name)
}

// GetEventFiltersPath gets the path of the event filter store.
func GetEventFiltersPath(ctx context.Context, name string) string {
	return eventFilterKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteEventFilterByName deletes an EventFilter by name.
func (s *Store) DeleteEventFilterByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name of filter")
	}

	resp, err := s.client.Delete(ctx, GetEventFiltersPath(ctx, name))
	if err != nil {
		return err
	}

	if resp.Deleted != 1 {
		return fmt.Errorf("filter %s does not exist", name)
	}

	return nil
}

// GetEventFilters gets the list of filters for a namespace.
func (s *Store) GetEventFilters(ctx context.Context, pred *store.SelectionPredicate) ([]*types.EventFilter, error) {
	filters := []*types.EventFilter{}
	err := List(ctx, s.client, GetEventFiltersPath, &filters, pred)
	return filters, err
}

// GetEventFilterByName gets an EventFilter by name.
func (s *Store) GetEventFilterByName(ctx context.Context, name string) (*types.EventFilter, error) {
	if name == "" {
		return nil, errors.New("must specify name of filter")
	}

	resp, err := s.client.Get(ctx, GetEventFiltersPath(ctx, name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	filterBytes := resp.Kvs[0].Value
	filter := &types.EventFilter{}
	if err := unmarshal(filterBytes, filter); err != nil {
		return nil, err
	}

	return filter, nil
}

// UpdateEventFilter updates an EventFilter.
func (s *Store) UpdateEventFilter(ctx context.Context, filter *types.EventFilter) error {
	if err := filter.Validate(); err != nil {
		return err
	}

	filterBytes, err := proto.Marshal(filter)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(filter.Namespace)), ">", 0)
	req := clientv3.OpPut(getEventFilterPath(filter), string(filterBytes))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the filter %s in namespace %s",
			filter.Name,
			filter.Namespace,
		)
	}

	return nil
}
