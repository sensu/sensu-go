package client

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

var rolesPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "roles")

// CreateRole with the given role
func (client *RestClient) CreateRole(role *types.Role) error {
	return client.Post(rolesPath(role.Namespace), role)
}

// DeleteRole with the given name
func (client *RestClient) DeleteRole(namespace, name string) error {
	return client.Delete(rolesPath(namespace, name))
}

// FetchRole with the given name
func (client *RestClient) FetchRole(name string) (*types.Role, error) {
	role := &types.Role{}
	if err := client.Get(rolesPath(client.config.Namespace(), name), role); err != nil {
		return nil, err
	}
	return role, nil
}

// ListRoles lists the roles within the given namespace.
func (client *RestClient) ListRoles(namespace string, options *ListOptions) ([]corev2.Role, error) {
	roles := []corev2.Role{}

	if err := client.List(rolesPath(namespace), &roles, options); err != nil {
		return roles, err
	}

	return roles, nil
}
