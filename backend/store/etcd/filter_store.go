package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

var (
	filtersPathPrefix = "filters"
	filterKeyBuilder  = newKeyBuilder(filtersPathPrefix)
)

func getFilterPath(filter *types.Filter) string {
	return filterKeyBuilder.withResource(filter).build(filter.Name)
}

func getFiltersPath(ctx context.Context, name string) string {
	return filterKeyBuilder.withContext(ctx).build(name)
}

func (s *etcdStore) DeleteFilterByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name of filter")
	}

	_, err := s.kvc.Delete(ctx, getFiltersPath(ctx, name))
	return err
}

// GetFilters gets the list of filters for an (optional) organization. Passing
// the empty string as the org will return all filters.
func (s *etcdStore) GetFilters(ctx context.Context) ([]*types.Filter, error) {
	resp, err := s.kvc.Get(ctx, getFiltersPath(ctx, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*types.Filter{}, nil
	}

	filtersArray := make([]*types.Filter, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		filter := &types.Filter{}
		err = json.Unmarshal(kv.Value, filter)
		if err != nil {
			return nil, err
		}
		filtersArray[i] = filter
	}

	return filtersArray, nil
}

func (s *etcdStore) GetFilterByName(ctx context.Context, name string) (*types.Filter, error) {
	if name == "" {
		return nil, errors.New("must specify name of filter")
	}

	resp, err := s.kvc.Get(ctx, getFiltersPath(ctx, name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	filterBytes := resp.Kvs[0].Value
	filter := &types.Filter{}
	if err := json.Unmarshal(filterBytes, filter); err != nil {
		return nil, err
	}

	return filter, nil
}

func (s *etcdStore) UpdateFilter(ctx context.Context, filter *types.Filter) error {
	if err := filter.Validate(); err != nil {
		return err
	}

	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getEnvironmentsPath(filter.Organization, filter.Environment)), ">", 0)
	req := clientv3.OpPut(getFilterPath(filter), string(filterBytes))
	res, err := s.kvc.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the filter %s in environment %s/%s",
			filter.Name,
			filter.Organization,
			filter.Environment,
		)
	}

	return nil
}
