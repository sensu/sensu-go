package client

import (
	"fmt"
	"net/url"
	"path"

	"github.com/sensu/sensu-go/types"
)

const roleBindingsBasePath = "/apis/rbac/v2/namespaces/%s/rolebindings"

func roleBindingsPath(namespace, name string) string {
	name = url.PathEscape(name)
	namespace = url.PathEscape(namespace)
	return path.Join(fmt.Sprintf(roleBindingsBasePath, namespace), name)
}

// CreateRoleBinding with the given role binding
func (client *RestClient) CreateRoleBinding(roleBinding *types.RoleBinding) error {
	return client.post(roleBindingsPath(roleBinding.Namespace, ""), roleBinding)
}

// DeleteRoleBinding with the given name
func (client *RestClient) DeleteRoleBinding(name string) error {
	return client.delete(roleBindingsPath(client.config.Namespace(), name))
}

// FetchRoleBinding with the given name
func (client *RestClient) FetchRoleBinding(name string) (*types.RoleBinding, error) {
	roleBinding := &types.RoleBinding{}
	if err := client.get(roleBindingsPath(client.config.Namespace(), name), roleBinding); err != nil {
		return nil, err
	}
	return roleBinding, nil
}

// ListRoleBindings lists the role bindings within the given namespace.
func (client *RestClient) ListRoleBindings(namespace string) ([]types.RoleBinding, error) {
	roleBindings := []types.RoleBinding{}

	if err := client.list(roleBindingsPath(namespace, ""), &roleBindings); err != nil {
		return roleBindings, err
	}

	return roleBindings, nil
}
