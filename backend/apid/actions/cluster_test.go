package actions

import (
	"testing"

	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
)

type mockCluster struct {
}

func (mockCluster) MemberList(context.Context) (*clientv3.MemberListResponse, error) {
	return new(clientv3.MemberListResponse), nil
}

func (mockCluster) MemberAdd(context.Context, []string) (*clientv3.MemberAddResponse, error) {
	return new(clientv3.MemberAddResponse), nil
}

func (mockCluster) MemberRemove(context.Context, uint64) (*clientv3.MemberRemoveResponse, error) {
	return new(clientv3.MemberRemoveResponse), nil
}

func (mockCluster) MemberUpdate(context.Context, uint64, []string) (*clientv3.MemberUpdateResponse, error) {
	return new(clientv3.MemberUpdateResponse), nil
}

var _ clientv3.Cluster = mockCluster{}

func TestMemberList(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})

	_, err := ctrl.MemberList(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemberAdd(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})

	_, err := ctrl.MemberAdd(context.Background(), []string{"foo"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemberUpdate(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})

	_, err := ctrl.MemberUpdate(context.Background(), 1234, []string{"foo"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemberRemove(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})

	_, err := ctrl.MemberRemove(context.Background(), 1234)
	if err != nil {
		t.Fatal(err)
	}
}
