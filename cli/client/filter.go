package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/sensu/sensu-go/types"
)

// CreateFilter creates a new filter on configured Sensu instance
func (client *RestClient) CreateFilter(filter *types.EventFilter) (err error) {
	bytes, err := json.Marshal(filter)
	if err != nil {
		return err
	}

	res, err := client.R().
		SetBody(bytes).
		Post("/filters")

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
	res, err := client.R().Delete("/filters/" + url.PathEscape(filter.Name))

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

	res, err := client.R().Get("/filters/" + url.PathEscape(name))
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
func (client *RestClient) ListFilters(org string) ([]types.EventFilter, error) {
	var filters []types.EventFilter
	res, err := client.R().Get("/filters?org=" + url.QueryEscape(org))
	if err != nil {
		return filters, err
	}

	if res.StatusCode() >= 400 {
		return filters, fmt.Errorf("%v", res.String())
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
	resp, err := client.R().SetBody(b).Put(fmt.Sprintf("/filters/%s", url.PathEscape(f.Name)))
	if err != nil {
		return err
	}

	if resp.StatusCode() >= 400 {
		err = errors.New(resp.String())
	}

	return err
}
