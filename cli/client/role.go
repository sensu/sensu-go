package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// CreateRole creates new role on configured Sensu instance
func (client *RestClient) CreateRole(role *types.Role) error {
	bytes, err := json.Marshal(role)
	if err != nil {
		return err
	}

	res, err := client.R().
		SetBody(bytes).
		Put("/rbac/roles")

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// DeleteRole deletes a role on configured Sensu instance
func (client *RestClient) DeleteRole(name string) error {
	res, err := client.R().Delete("/rbac/roles/" + name)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// ListRoles fetches all roles from configured Sensu instance
func (client *RestClient) ListRoles() ([]types.Role, error) {
	var roles []types.Role

	res, err := client.R().Get("/rbac/roles")
	if err != nil {
		return roles, err
	}

	if res.StatusCode() >= 400 {
		return roles, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &roles)
	return roles, err
}

// AddRule adds new rule to existing role for configured Sensu instance
func (client *RestClient) AddRule(roleName string, rule *types.Rule) error {
	bytes, err := json.Marshal(rule)
	if err != nil {
		return err
	}

	key := "/rbac/roles/" + roleName + "/type/" + rule.Type
	res, err := client.R().SetBody(bytes).Put(key)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// RemoveRule removes rule from existing role for configured Sensu instance
func (client *RestClient) RemoveRule(name string, t string) error {
	key := "/rbac/roles/" + name + "/types/" + t
	res, err := client.R().Delete(key)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}
