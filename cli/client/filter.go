package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

var filtersPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "filters")

// CreateFilter creates a new filter on configured Sensu instance
func (client *RestClient) CreateFilter(filter *types.EventFilter) (err error) {
	bytes, err := json.Marshal(filter)
	if err != nil {
		return err
	}

	path := filtersPath(client.config.Namespace())
	res, err := client.R().SetBody(bytes).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// DeleteFilter deletes a filter from configured Sensu instance
func (client *RestClient) DeleteFilter(filter *types.EventFilter) error {
	path := filtersPath(filter.Namespace, filter.Name)
	res, err := client.R().Delete(path)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// FetchFilter fetches a specific check
func (client *RestClient) FetchFilter(name string) (*types.EventFilter, error) {
	var filter *types.EventFilter

	path := filtersPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &filter)
	return filter, err
}

// ListFilters fetches all filters from configured Sensu instance
func (client *RestClient) ListFilters(namespace string) ([]types.EventFilter, error) {
	var filters []types.EventFilter

	path := filtersPath(namespace)
	res, err := client.R().Get(path)
	if err != nil {
		return filters, err
	}

	if res.StatusCode() >= 400 {
		return filters, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &filters)
	return filters, err
}

// UpdateFilter updates an existing filter with fields from a new one.
func (client *RestClient) UpdateFilter(f *types.EventFilter) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}

	path := filtersPath(f.Namespace, f.Name)
	resp, err := client.R().SetBody(b).Put(path)
	if err != nil {
		return err
	}

	if resp.StatusCode() >= 400 {
		err = errors.New(resp.String())
	}

	return err
}
