package client

import (
	"encoding/json"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/core/v3/types"
)

// MutatorsPath is the api path for mutators.
var MutatorsPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "mutators")

// CreateMutator creates new mutator on the configured Sensu instance
func (client *RestClient) CreateMutator(mutator *corev2.Mutator) (err error) {
	bytes, err := json.Marshal(types.WrapResource(mutator))
	if err != nil {
		return err
	}

	path := MutatorsPath(mutator.Namespace)
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
	return client.Delete(MutatorsPath(namespace, name))
}

// FetchMutator fetches a specific handler from the configured Sensu instance
func (client *RestClient) FetchMutator(name string) (*corev2.Mutator, error) {
	path := MutatorsPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	var wrapper types.Wrapper
	err = json.Unmarshal(res.Body(), &wrapper)
	return wrapper.Value.(*corev2.Mutator), err
}

// UpdateMutator updates a given mutator on a configured Sensu instance
func (client *RestClient) UpdateMutator(mutator *corev2.Mutator) (err error) {
	bytes, err := json.Marshal(types.WrapResource(mutator))
	if err != nil {
		return err
	}

	path := MutatorsPath(mutator.Namespace, mutator.Name)
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
