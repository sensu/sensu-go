package client

import (
	"net/url"
	"path"

	"github.com/sensu/sensu-go/types"
)

const clusterRoleBindingsBasePath = "/apis/rbac/v2/clusterrolebindings"

func clusterRoleBindingsPath(name string) string {
	name = url.PathEscape(name)
	return path.Join(clusterRoleBindingsBasePath, name)
}

// CreateClusterRoleBinding with the given cluster role binding
func (client *RestClient) CreateClusterRoleBinding(clusterRoleBinding *types.ClusterRoleBinding) error {
	return client.post(clusterRoleBindingsBasePath, clusterRoleBinding)
}

// DeleteClusterRoleBinding with the given name
func (client *RestClient) DeleteClusterRoleBinding(name string) error {
	return client.delete(clusterRoleBindingsPath(name))
}

// FetchClusterRoleBinding with the given name
func (client *RestClient) FetchClusterRoleBinding(name string) (*types.ClusterRoleBinding, error) {
	clusterRoleBinding := &types.ClusterRoleBinding{}
	if err := client.get(clusterRoleBindingsPath(name), clusterRoleBinding); err != nil {
		return nil, err
	}
	return clusterRoleBinding, nil
}

// ListClusterRoleBinding in the cluster
func (client *RestClient) ListClusterRoleBindings() ([]types.ClusterRoleBinding, error) {
	clusterRoleBindings := []types.ClusterRoleBinding{}

	if err := client.list(clusterRoleBindingsBasePath, &clusterRoleBindings); err != nil {
		return nil, err
	}

	return clusterRoleBindings, nil
}
