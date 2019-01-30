package client

import (
	"github.com/sensu/sensu-go/types"
)

var clusterRolesPath = CreateBasePath(coreAPIGroup, coreAPIVersion, "clusterroles")

// CreateClusterRole with the given cluster role
func (client *RestClient) CreateClusterRole(clusterRole *types.ClusterRole) error {
	return client.Post(clusterRolesPath(), clusterRole)
}

// DeleteClusterRole with the given name
func (client *RestClient) DeleteClusterRole(name string) error {
	return client.Delete(clusterRolesPath(name))
}

// FetchClusterRole with the given name
func (client *RestClient) FetchClusterRole(name string) (*types.ClusterRole, error) {
	clusterRole := &types.ClusterRole{}
	if err := client.Get(clusterRolesPath(name), clusterRole); err != nil {
		return nil, err
	}
	return clusterRole, nil
}

// ListClusterRoles within the namespace
func (client *RestClient) ListClusterRoles() ([]types.ClusterRole, error) {
	clusterRoles := []types.ClusterRole{}

	if err := client.List(clusterRolesPath(), &clusterRoles); err != nil {
		return clusterRoles, err
	}

	return clusterRoles, nil
}
