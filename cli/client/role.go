package client

import (
	"github.com/sensu/sensu-go/types"
)

var rolesPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "roles")

// CreateRole with the given role
func (client *RestClient) CreateRole(role *types.Role) error {
	return client.post(rolesPath(role.Namespace), role)
}

// DeleteRole with the given name
func (client *RestClient) DeleteRole(name string) error {
	return client.delete(rolesPath(client.config.Namespace(), name))
}

// FetchRole with the given name
func (client *RestClient) FetchRole(name string) (*types.Role, error) {
	role := &types.Role{}
	if err := client.get(rolesPath(client.config.Namespace(), name), role); err != nil {
		return nil, err
	}
	return role, nil
}

// ListRoles lists the roles within the given namespace.
func (client *RestClient) ListRoles(namespace string) ([]types.Role, error) {
	roles := []types.Role{}

	if err := client.list(rolesPath(namespace), &roles); err != nil {
		return roles, err
	}

	return roles, nil
}
