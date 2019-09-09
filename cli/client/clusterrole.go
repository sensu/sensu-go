package client

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// ClusterRolesPath is the api path for cluster roles.
var ClusterRolesPath = CreateBasePath(coreAPIGroup, coreAPIVersion, "clusterroles")

// CreateClusterRole with the given cluster role
func (client *RestClient) CreateClusterRole(clusterRole *corev2.ClusterRole) error {
	return client.Post(ClusterRolesPath(), clusterRole)
}

// DeleteClusterRole with the given name
func (client *RestClient) DeleteClusterRole(name string) error {
	return client.Delete(ClusterRolesPath(name))
}

// FetchClusterRole with the given name
func (client *RestClient) FetchClusterRole(name string) (*corev2.ClusterRole, error) {
	clusterRole := &corev2.ClusterRole{}
	if err := client.Get(ClusterRolesPath(name), clusterRole); err != nil {
		return nil, err
	}
	return clusterRole, nil
}
