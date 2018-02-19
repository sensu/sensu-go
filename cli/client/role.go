package client

import (
	"encoding/json"
	"net/url"
	"path"

	"github.com/sensu/sensu-go/types"
)

const rolesBasePath = "/rbac/roles"

func rolesPath(ext ...string) string {
	parts := ext
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return path.Join(append([]string{rolesBasePath}, parts...)...)
}

// CreateRole creates new role on configured Sensu instance
func (client *RestClient) CreateRole(role *types.Role) error {
	res, err := client.R().SetBody(role).Post(rolesBasePath)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return unmarshalError(res)
	}

	return nil
}

// DeleteRole deletes a role on configured Sensu instance
func (client *RestClient) DeleteRole(name string) error {
	res, err := client.R().Delete(rolesPath(name))
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return unmarshalError(res)
	}

	return nil
}

// FetchRole fetches role from configured Sensu instance
func (client *RestClient) FetchRole(name string) (*types.Role, error) {
	var role types.Role

	res, cerr := client.R().SetResult(&role).Get(rolesPath(name))
	if cerr != nil {
		return nil, cerr
	}

	if res.StatusCode() >= 400 {
		return nil, unmarshalError(res)
	}

	return &role, nil
}

// ListRoles fetches all roles from configured Sensu instance
func (client *RestClient) ListRoles() ([]types.Role, error) {
	var roles []types.Role

	res, err := client.R().Get("/rbac/roles")
	if err != nil {
		return roles, err
	}

	if res.StatusCode() >= 400 {
		return roles, unmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &roles)
	return roles, err
}

// AddRule adds new rule to existing role for configured Sensu instance
func (client *RestClient) AddRule(roleName string, rule *types.Rule) error {
	key := rolesPath(roleName, "rules", rule.Type)
	res, err := client.R().SetBody(rule).Put(key)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return unmarshalError(res)
	}

	return nil
}

// RemoveRule removes rule from existing role for configured Sensu instance
func (client *RestClient) RemoveRule(name string, t string) error {
	path := rolesPath(name, "rules", t)
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return unmarshalError(res)
	}

	return nil
}
