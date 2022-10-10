package client

import (
	corev2 "github.com/sensu/core/v2"
)

// ClusterRoleBindingsPath is the api path for cluster role bindings.
var ClusterRoleBindingsPath = CreateBasePath(coreAPIGroup, coreAPIVersion, "clusterrolebindings")

// CreateClusterRoleBinding with the given cluster role binding
func (client *RestClient) CreateClusterRoleBinding(clusterRoleBinding *corev2.ClusterRoleBinding) error {
	return client.Post(ClusterRoleBindingsPath(), clusterRoleBinding)
}

// DeleteClusterRoleBinding with the given name
func (client *RestClient) DeleteClusterRoleBinding(name string) error {
	return client.Delete(ClusterRoleBindingsPath(name))
}

// FetchClusterRoleBinding with the given name
func (client *RestClient) FetchClusterRoleBinding(name string) (*corev2.ClusterRoleBinding, error) {
	clusterRoleBinding := &corev2.ClusterRoleBinding{}
	if err := client.Get(ClusterRoleBindingsPath(name), clusterRoleBinding); err != nil {
		return nil, err
	}
	return clusterRoleBinding, nil
}
