package actions

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/authorization"
)

// ClusterController is a thin wrapper around clientv3.Cluster. It exists
// only for the purposes of access control.
type ClusterController struct {
	cluster clientv3.Cluster
	policy  authorization.ClusterPolicy
}

func NewClusterController(cluster clientv3.Cluster) ClusterController {
	return ClusterController{
		cluster: cluster,
		policy:  authorization.ClusterPolicy{},
	}
}

func (c ClusterController) MemberList(ctx context.Context) (*clientv3.MemberListResponse, error) {
	abilities := c.policy.WithContext(ctx)
	if !abilities.HasPermission() {
		return nil, NewErrorf(PermissionDenied)
	}
	return c.cluster.MemberList(ctx)
}

func (c ClusterController) MemberAdd(ctx context.Context, addrs []string) (*clientv3.MemberAddResponse, error) {
	abilities := c.policy.WithContext(ctx)
	if !abilities.HasPermission() {
		return nil, NewErrorf(PermissionDenied)
	}
	return c.cluster.MemberAdd(ctx, addrs)
}

func (c ClusterController) MemberRemove(ctx context.Context, id uint64) (*clientv3.MemberRemoveResponse, error) {
	abilities := c.policy.WithContext(ctx)
	if !abilities.HasPermission() {
		return nil, NewErrorf(PermissionDenied)
	}
	return c.cluster.MemberRemove(ctx, id)
}

func (c ClusterController) MemberUpdate(ctx context.Context, id uint64, addrs []string) (*clientv3.MemberUpdateResponse, error) {
	abilities := c.policy.WithContext(ctx)
	if !abilities.HasPermission() {
		return nil, NewErrorf(PermissionDenied)
	}
	return c.cluster.MemberUpdate(ctx, id, addrs)
}
