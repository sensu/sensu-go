package client

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sensu/sensu-go/types"
)

// ListMutators fetches all mutators from configured Sensu instance
func (client *RestClient) ListMutators() ([]types.Mutator, error) {
	var mutators []types.Mutator

	res, err := client.R().Get("/mutators")
	if err != nil {
		return mutators, err
	}

	if res.StatusCode() >= 400 {
		return mutators, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &mutators)
	return mutators, err
}

// CreateMutator creates new mutator on configured Sensu instance
func (client *RestClient) CreateMutator(mutator *types.Mutator) (err error) {
	bytes, err := json.Marshal(mutator)
	if err == nil {
		_, err = client.R().
			SetBody(bytes).
			Put("/mutators/" + url.PathEscape(mutator.Name))
	}

	return
}
