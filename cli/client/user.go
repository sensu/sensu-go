package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// CreateUser creates new check on configured Sensu instance
func (client *RestClient) CreateUser(user *types.User) error {
	bytes, err := json.Marshal(user)
	if err != nil {
		return err
	}

	res, err := client.R().
		SetBody(bytes).
		Put("/users")

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}
