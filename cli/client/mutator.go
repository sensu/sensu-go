package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

var mutatorsPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "mutators")

// ListMutators fetches all mutators from the configured Sensu instance
func (client *RestClient) ListMutators(namespace string) ([]types.Mutator, error) {
	var mutators []types.Mutator

	path := mutatorsPath(namespace)
	res, err := client.R().Get(path)
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

	path := mutatorsPath(mutator.Namespace)
	res, err := client.R().SetBody(bytes).Post(path)
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
	path := mutatorsPath(client.config.Namespace(), mutator.Name)
	res, err := client.R().Delete(path)
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

	path := mutatorsPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
	if err != nil {
		return mutator, err
	}

	if res.StatusCode() >= 400 {
		return mutator, UnmarshalError(res)
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

	path := mutatorsPath(mutator.Namespace, mutator.Name)
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
