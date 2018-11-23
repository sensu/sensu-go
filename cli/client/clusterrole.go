package client

import (
	"net/url"
	"path"

	"github.com/sensu/sensu-go/types"
)

const clusterRolesBasePath = "/apis/rbac/v2/clusterroles"

func clusterRolesPath(name string) string {
	name = url.PathEscape(name)
	return path.Join(clusterRolesBasePath, name)
}

// CreateClusterRole with the given cluster role
func (client *RestClient) CreateClusterRole(clusterRole *types.ClusterRole) error {
	return client.post(clusterRolesBasePath, clusterRole)
}

// DeleteClusterRole with the given name
func (client *RestClient) DeleteClusterRole(name string) error {
	return client.delete(clusterRolesPath(name))
}

// FetchClusterRole with the given name
func (client *RestClient) FetchClusterRole(name string) (*types.ClusterRole, error) {
	clusterRole := &types.ClusterRole{}
	if err := client.get(clusterRolesPath(name), clusterRole); err != nil {
		return nil, err
	}
	return clusterRole, nil
}

// ListClusterRoles within the namespace
func (client *RestClient) ListClusterRoles() ([]types.ClusterRole, error) {
	clusterRoles := []types.ClusterRole{}

	if err := client.list(clusterRolesBasePath, &clusterRoles); err != nil {
		return clusterRoles, err
	}

	return clusterRoles, nil
}
