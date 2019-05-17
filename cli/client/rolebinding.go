package client

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

var roleBindingsPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "rolebindings")

// CreateRoleBinding with the given role binding
func (client *RestClient) CreateRoleBinding(roleBinding *types.RoleBinding) error {
	return client.Post(roleBindingsPath(roleBinding.Namespace), roleBinding)
}

// DeleteRoleBinding with the given name
func (client *RestClient) DeleteRoleBinding(namespace, name string) error {
	return client.Delete(roleBindingsPath(namespace, name))
}

// FetchRoleBinding with the given name
func (client *RestClient) FetchRoleBinding(name string) (*types.RoleBinding, error) {
	roleBinding := &types.RoleBinding{}
	if err := client.Get(roleBindingsPath(client.config.Namespace(), name), roleBinding); err != nil {
		return nil, err
	}
	return roleBinding, nil
}

// ListRoleBindings lists the role bindings within the given namespace.
func (client *RestClient) ListRoleBindings(namespace string, options *ListOptions) ([]corev2.RoleBinding, error) {
	roleBindings := []corev2.RoleBinding{}

	if err := client.List(roleBindingsPath(namespace), &roleBindings, options); err != nil {
		return roleBindings, err
	}

	return roleBindings, nil
}
