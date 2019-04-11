package actions

import (
	"context"

	"github.com/coreos/etcd/clientv3"
)

// ClusterController is a thin wrapper around clientv3.Cluster. It exists
// only for the purposes of access control.
type ClusterController struct {
	cluster clientv3.Cluster
}

// NewClusterController provides a new controller for the etcd cluster
func NewClusterController(cluster clientv3.Cluster) ClusterController {
	return ClusterController{
		cluster: cluster,
	}
}

// MemberList returns the list of members in the cluster
func (c ClusterController) MemberList(ctx context.Context) (*clientv3.MemberListResponse, error) {
	return c.cluster.MemberList(ctx)
}

// MemberAdd adds a member to the cluster
func (c ClusterController) MemberAdd(ctx context.Context, addrs []string) (*clientv3.MemberAddResponse, error) {
	return c.cluster.MemberAdd(ctx, addrs)
}

// MemberRemove removes a member from the cluster
func (c ClusterController) MemberRemove(ctx context.Context, id uint64) (*clientv3.MemberRemoveResponse, error) {
	return c.cluster.MemberRemove(ctx, id)
}

// MemberUpdate updates the specified member addresses
func (c ClusterController) MemberUpdate(ctx context.Context, id uint64, addrs []string) (*clientv3.MemberUpdateResponse, error) {
	return c.cluster.MemberUpdate(ctx, id, addrs)
}
