package client

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
func (client *RestClient) ListClusterRoles(options ListOptions) ([]corev2.ClusterRole, string, error) {
	var header string
	clusterRoles := []corev2.ClusterRole{}

	header, err := client.List(clusterRolesPath(), &clusterRoles, options)
	if err != nil {
		return clusterRoles, header, err
	}

	return clusterRoles, header, nil
}
