package client

import (
	"encoding/json"

	"github.com/sensu/sensu-go/types"
)

// ListChecks fetches all checks from configured Sensu instance
func (client *RestClient) ListChecks() (checks []types.Check, err error) {
	r, err := client.R().Get("/checks")
	if err == nil {
		err = json.Unmarshal(r.Body(), &checks)
	}

	return
}

// CreateCheck creates new check on configured Sensu instance
func (client *RestClient) CreateCheck(check *types.Check) (err error) {
	bytes, err := json.Marshal(check)
	if err == nil {
		_, err = client.R().
			SetBody(bytes).
			Put("/checks/" + check.Name)
	}

	return
}

// DeleteCheck deletes check from configured Sensu instance
func (client *RestClient) DeleteCheck(check *types.Check) (err error) {
	_, err = client.R().Delete("/checks/" + check.Name)
	return
}
