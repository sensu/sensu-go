package client

import (
	"encoding/json"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

var mutatorsPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "mutators")

// ListMutators fetches all mutators from the configured Sensu instance
func (client *RestClient) ListMutators(namespace string, options ListOptions) ([]corev2.Mutator, error) {
	var mutators []corev2.Mutator

	path := mutatorsPath(namespace)
	request := client.R()

	ApplyListOptions(request, options)

	res, err := request.Get(path)
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
func (client *RestClient) DeleteMutator(namespace, name string) (err error) {
	return client.Delete(mutatorsPath(namespace, name))
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
