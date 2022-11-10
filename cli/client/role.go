package client

import (
	corev2 "github.com/sensu/core/v2"
)

// RolesPath is the api path for roles.
var RolesPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "roles")

// CreateRole with the given role
func (client *RestClient) CreateRole(role *corev2.Role) error {
	return client.Post(RolesPath(role.Namespace), role)
}

// DeleteRole with the given name
func (client *RestClient) DeleteRole(namespace, name string) error {
	return client.Delete(RolesPath(namespace, name))
}

// FetchRole with the given name
func (client *RestClient) FetchRole(name string) (*corev2.Role, error) {
	role := &corev2.Role{}
	if err := client.Get(RolesPath(client.config.Namespace(), name), role); err != nil {
		return nil, err
	}
	return role, nil
}
