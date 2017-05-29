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
		Put("/rbac/users")

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// DeleteUser deletes a user on configured Sensu instance
func (client *RestClient) DeleteUser(username string) error {
	res, err := client.R().Delete("/rbac/users/" + username)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// ListUsers fetches all users from configured Sensu instance
func (client *RestClient) ListUsers() ([]types.User, error) {
	var users []types.User

	res, err := client.R().Get("/rbac/users")
	if err != nil {
		return users, err
	}

	if res.StatusCode() >= 400 {
		return users, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &users)
	return users, err
}
