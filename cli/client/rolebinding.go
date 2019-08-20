package client

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// RoleBindingsPath is the api path for role bindings.
var RoleBindingsPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "rolebindings")

// CreateRoleBinding with the given role binding
func (client *RestClient) CreateRoleBinding(roleBinding *corev2.RoleBinding) error {
	return client.Post(RoleBindingsPath(roleBinding.Namespace), roleBinding)
}

// DeleteRoleBinding with the given name
func (client *RestClient) DeleteRoleBinding(namespace, name string) error {
	return client.Delete(RoleBindingsPath(namespace, name))
}

// FetchRoleBinding with the given name
func (client *RestClient) FetchRoleBinding(name string) (*corev2.RoleBinding, error) {
	roleBinding := &corev2.RoleBinding{}
	if err := client.Get(RoleBindingsPath(client.config.Namespace(), name), roleBinding); err != nil {
		return nil, err
	}
	return roleBinding, nil
}
