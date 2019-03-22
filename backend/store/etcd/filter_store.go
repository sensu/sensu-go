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
	eventFiltersPathPrefix = "event-filters"
	eventFilterKeyBuilder  = store.NewKeyBuilder(eventFiltersPathPrefix)
)

func getEventFilterPath(filter *types.EventFilter) string {
	return eventFilterKeyBuilder.WithResource(filter).Build(filter.Name)
}

func getEventFiltersPath(ctx context.Context, name string) string {
	return eventFilterKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteEventFilterByName deletes an EventFilter by name.
func (s *Store) DeleteEventFilterByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name of filter")
	}

	resp, err := s.client.Delete(ctx, getEventFiltersPath(ctx, name))
	if err != nil {
		return err
	}

	if resp.Deleted != 1 {
		return fmt.Errorf("filter %s does not exist", name)
	}

	return nil
}

// GetEventFilters gets the list of filters for a namespace.
func (s *Store) GetEventFilters(ctx context.Context, pageSize int64, continueToken string) (filters []*types.EventFilter, newContinueToken string, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithLimit(pageSize),
	}

	keyPrefix := getEventFiltersPath(ctx, "")
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	resp, err := s.client.Get(ctx, path.Join(keyPrefix, continueToken), opts...)
	if err != nil {
		return nil, "", err
	}
	if len(resp.Kvs) == 0 {
		return []*types.EventFilter{}, "", nil
	}

	for _, kv := range resp.Kvs {
		filter := &types.EventFilter{}
		err = json.Unmarshal(kv.Value, filter)
		if err != nil {
			return nil, "", err
		}

		filters = append(filters, filter)
	}

	if pageSize != 0 && resp.Count > pageSize {
		lastFilter := filters[len(filters)-1]
		newContinueToken = lastFilter.Name + "\x00"
	}

	return filters, newContinueToken, nil
}

// GetEventFilterByName gets an EventFilter by name.
func (s *Store) GetEventFilterByName(ctx context.Context, name string) (*types.EventFilter, error) {
	if name == "" {
		return nil, errors.New("must specify name of filter")
	}

	resp, err := s.client.Get(ctx, getEventFiltersPath(ctx, name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	filterBytes := resp.Kvs[0].Value
	filter := &types.EventFilter{}
	if err := json.Unmarshal(filterBytes, filter); err != nil {
		return nil, err
	}

	return filter, nil
}

// UpdateEventFilter updates an EventFilter.
func (s *Store) UpdateEventFilter(ctx context.Context, filter *types.EventFilter) error {
	if err := filter.Validate(); err != nil {
		return err
	}

	filterBytes, err := json.Marshal(filter)
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
