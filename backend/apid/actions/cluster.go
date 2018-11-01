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

func NewClusterController(cluster clientv3.Cluster) ClusterController {
	return ClusterController{
		cluster: cluster,
	}
}

func (c ClusterController) MemberList(ctx context.Context) (*clientv3.MemberListResponse, error) {
	return c.cluster.MemberList(ctx)
}

func (c ClusterController) MemberAdd(ctx context.Context, addrs []string) (*clientv3.MemberAddResponse, error) {
	return c.cluster.MemberAdd(ctx, addrs)
}

func (c ClusterController) MemberRemove(ctx context.Context, id uint64) (*clientv3.MemberRemoveResponse, error) {
	return c.cluster.MemberRemove(ctx, id)
}

func (c ClusterController) MemberUpdate(ctx context.Context, id uint64, addrs []string) (*clientv3.MemberUpdateResponse, error) {
	return c.cluster.MemberUpdate(ctx, id, addrs)
}
