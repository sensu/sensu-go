package client

import (
	"encoding/json"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// FiltersPath is the api path for filters.
var FiltersPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "filters")

// CreateFilter creates a new filter on configured Sensu instance
func (client *RestClient) CreateFilter(filter *corev2.EventFilter) (err error) {
	bytes, err := json.Marshal(filter)
	if err != nil {
		return err
	}

	path := FiltersPath(client.config.Namespace())
	res, err := client.R().SetBody(bytes).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// DeleteFilter deletes a filter from configured Sensu instance
func (client *RestClient) DeleteFilter(namespace, name string) error {
	return client.Delete(FiltersPath(namespace, name))
}

// FetchFilter fetches a specific check
func (client *RestClient) FetchFilter(name string) (*corev2.EventFilter, error) {
	var filter *corev2.EventFilter

	path := FiltersPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &filter)
	return filter, err
}

// UpdateFilter updates an existing filter with fields from a new one.
func (client *RestClient) UpdateFilter(f *corev2.EventFilter) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}

	path := FiltersPath(f.Namespace, f.Name)
	resp, err := client.R().SetBody(b).Put(path)
	if err != nil {
		return err
	}

	if resp.StatusCode() >= 400 {
		return UnmarshalError(resp)
	}

	return err
}
