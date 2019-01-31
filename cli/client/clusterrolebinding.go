package client

import (
	"github.com/sensu/sensu-go/types"
)

var clusterRoleBindingsPath = CreateBasePath(coreAPIGroup, coreAPIVersion, "clusterrolebindings")

// CreateClusterRoleBinding with the given cluster role binding
func (client *RestClient) CreateClusterRoleBinding(clusterRoleBinding *types.ClusterRoleBinding) error {
	return client.Post(clusterRoleBindingsPath(), clusterRoleBinding)
}

// DeleteClusterRoleBinding with the given name
func (client *RestClient) DeleteClusterRoleBinding(name string) error {
	return client.Delete(clusterRoleBindingsPath(name))
}

// FetchClusterRoleBinding with the given name
func (client *RestClient) FetchClusterRoleBinding(name string) (*types.ClusterRoleBinding, error) {
	clusterRoleBinding := &types.ClusterRoleBinding{}
	if err := client.Get(clusterRoleBindingsPath(name), clusterRoleBinding); err != nil {
		return nil, err
	}
	return clusterRoleBinding, nil
}

// ListClusterRoleBinding in the cluster
func (client *RestClient) ListClusterRoleBindings() ([]types.ClusterRoleBinding, error) {
	clusterRoleBindings := []types.ClusterRoleBinding{}

	if err := client.List(clusterRoleBindingsPath(), &clusterRoleBindings); err != nil {
		return clusterRoleBindings, err
	}

	return clusterRoleBindings, nil
}
