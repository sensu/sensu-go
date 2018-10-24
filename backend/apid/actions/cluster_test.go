package actions

import (
	"context"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
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

func TestMemberListNoPerm(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})
	_, err := ctrl.MemberList(context.TODO())
	if err == nil {
		t.Fatal("expected error")
	}
	e := err.(Error)
	if got, want := e.Code, PermissionDenied; got != want {
		t.Errorf("bad error code: got %v, want %v", got, want)
	}
}

func TestMemberAddNoPerm(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})
	_, err := ctrl.MemberAdd(context.TODO(), []string{"foo"})
	if err == nil {
		t.Fatal("expected error")
	}
	e := err.(Error)
	if got, want := e.Code, PermissionDenied; got != want {
		t.Errorf("bad error code: got %v, want %v", got, want)
	}
}

func TestMemberRemoveNoPerm(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})
	_, err := ctrl.MemberRemove(context.TODO(), 1234)
	if err == nil {
		t.Fatal("expected error")
	}
	e := err.(Error)
	if got, want := e.Code, PermissionDenied; got != want {
		t.Errorf("bad error code: got %v, want %v", got, want)
	}
}

func TestMemberUpdateNoPerm(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})
	_, err := ctrl.MemberUpdate(context.TODO(), 1234, []string{"foo"})
	if err == nil {
		t.Fatal("expected error")
	}
	e := err.(Error)
	if got, want := e.Code, PermissionDenied; got != want {
		t.Errorf("bad error code: got %v, want %v", got, want)
	}
}

func TestMemberList(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})
	ctx := testutil.NewContext(
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeCluster, types.RuleAllPerms...)))

	_, err := ctrl.MemberList(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemberAdd(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})
	ctx := testutil.NewContext(
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeCluster, types.RuleAllPerms...)))

	_, err := ctrl.MemberAdd(ctx, []string{"foo"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemberUpdate(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})
	ctx := testutil.NewContext(
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeCluster, types.RuleAllPerms...)))

	_, err := ctrl.MemberUpdate(ctx, 1234, []string{"foo"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemberRemove(t *testing.T) {
	ctrl := NewClusterController(mockCluster{})
	ctx := testutil.NewContext(
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeCluster, types.RuleAllPerms...)))

	_, err := ctrl.MemberRemove(ctx, 1234)
	if err != nil {
		t.Fatal(err)
	}
}
