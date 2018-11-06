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
	return path.Join(fmt.Sprintf(rolesBasePath, namespace), name)
}

// CreateRoleBinding with the given role binding
func (client *RestClient) CreateRoleBinding(roleBinding *types.RoleBinding) error {
	return client.post(rolesPath(roleBinding.Namespace, ""), roleBinding)
}

// DeleteRoleBinding with the given name
func (client *RestClient) DeleteRoleBinding(name string) error {
	return client.delete(rolesPath(client.config.Namespace(), name))
}

// FetchRoleBinding with the given name
func (client *RestClient) FetchRoleBinding(name string) (*types.RoleBinding, error) {
	roleBinding := &types.RoleBinding{}
	if err := client.get(rolesPath(client.config.Namespace(), name), roleBinding); err != nil {
		return nil, err
	}
	return roleBinding, nil
}

// ListRoleBindings within the namespace
func (client *RestClient) ListRoleBindings() ([]types.RoleBinding, error) {
	roleBindings := []types.RoleBinding{}

	if err := client.list(rolesPath(client.config.Namespace(), ""), &roleBindings); err != nil {
		return nil, err
	}

	return roleBindings, nil
}
