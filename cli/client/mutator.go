package client

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sensu/sensu-go/types"
)

// ListMutators fetches all mutators from the configured Sensu instance
func (client *RestClient) ListMutators(namespace string) ([]types.Mutator, error) {
	var mutators []types.Mutator

	res, err := client.R().Get("/mutators?namespace=" + url.QueryEscape(namespace))
	if err != nil {
		return mutators, err
	}

	if res.StatusCode() >= 400 {
		return mutators, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &mutators)
	return mutators, err
}

// CreateMutator creates new mutator on the configured Sensu instance
func (client *RestClient) CreateMutator(mutator *types.Mutator) (err error) {
	bytes, err := json.Marshal(mutator)
	if err != nil {
		return err

	}
	res, err := client.R().SetBody(bytes).Post("/mutators")
	if err != nil {
		return err

	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}
	return nil
}

// DeleteMutator deletes the given mutator from the configured Sensu instance
func (client *RestClient) DeleteMutator(mutator *types.Mutator) (err error) {
	res, err := client.R().Delete("/mutators/" + url.PathEscape(mutator.Name))
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// FetchMutator fetches a specific handler from the configured Sensu instance
func (client *RestClient) FetchMutator(name string) (*types.Mutator, error) {
	var mutator *types.Mutator

	res, err := client.R().Get("/mutators/" + url.PathEscape(name))
	if err != nil {
		return mutator, err
	}

	if res.StatusCode() >= 400 {
		return mutator, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &mutator)
	return mutator, err
}

// UpdateMutator updates a given mutator on a configured Sensu instance
func (client *RestClient) UpdateMutator(mutator *types.Mutator) (err error) {
	bytes, err := json.Marshal(mutator)
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Put("/mutators/" + url.PathEscape(mutator.Name))
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
