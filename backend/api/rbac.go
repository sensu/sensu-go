package api

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type RBACClient struct {
	roleC               GenericClient
	roleBindingC        GenericClient
	clusterRoleC        GenericClient
	clusterRoleBindingC GenericClient
	auth                authorization.Authorizer
}

func NewRBACClient(store storev2.Interface, auth authorization.Authorizer) *RBACClient {
	return &RBACClient{
		auth: auth,
		roleC: GenericClient{
			Kind:       &corev2.Role{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v2",
		},
		roleBindingC: GenericClient{
			Kind:       &corev2.RoleBinding{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v2",
		},
		clusterRoleC: GenericClient{
			Kind:       &corev2.ClusterRole{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v2",
		},
		clusterRoleBindingC: GenericClient{
			Kind:       &corev2.ClusterRoleBinding{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v2",
		},
	}
}

// ListRoleBindings fetches a list of role binding resources, if authorized.
func (a *RBACClient) ListRoleBindings(ctx context.Context) ([]*corev2.RoleBinding, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.RoleBinding{}
	if err := a.roleBindingC.List(ctx, &slice, pred); err != nil {
		return nil, err
	}
	return slice, nil
}

// FetchRoleBinding fetches an role binding resource from the backend, if authorized.
func (a *RBACClient) FetchRoleBinding(ctx context.Context, name string) (*corev2.RoleBinding, error) {
	var rb corev2.RoleBinding
	if err := a.roleBindingC.Get(ctx, name, &rb); err != nil {
		return nil, err
	}
	return &rb, nil
}

// CreateRoleBinding creates an role binding resource, if authorized.
func (a *RBACClient) CreateRoleBinding(ctx context.Context, rb *corev2.RoleBinding) error {
	if err := a.roleBindingC.Create(ctx, rb); err != nil {
		return err
	}
	return nil
}

// UpdateRoleBinding updates an role binding resource, if authorized.
func (a *RBACClient) UpdateRoleBinding(ctx context.Context, rb *corev2.RoleBinding) error {
	if err := a.roleBindingC.Update(ctx, rb); err != nil {
		return err
	}
	return nil
}

// ListRoles fetches a list of role resources, if authorized.
func (a *RBACClient) ListRoles(ctx context.Context) ([]*corev2.Role, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.Role{}
	if err := a.roleC.List(ctx, &slice, pred); err != nil {
		return nil, err
	}
	return slice, nil
}

// FetchRole fetches an role resource from the backend, if authorized.
func (a *RBACClient) FetchRole(ctx context.Context, name string) (*corev2.Role, error) {
	var rb corev2.Role
	if err := a.roleC.Get(ctx, name, &rb); err != nil {
		return nil, err
	}
	return &rb, nil
}

// CreateRole creates an role resource, if authorized.
func (a *RBACClient) CreateRole(ctx context.Context, rb *corev2.Role) error {
	if err := a.roleC.Create(ctx, rb); err != nil {
		return err
	}
	return nil
}

// UpdateRole updates an role resource, if authorized.
func (a *RBACClient) UpdateRole(ctx context.Context, rb *corev2.Role) error {
	if err := a.roleC.Update(ctx, rb); err != nil {
		return err
	}
	return nil
}

// ListClusterRoleBindings fetches a list of role binding resources, if authorized.
func (a *RBACClient) ListClusterRoleBindings(ctx context.Context) ([]*corev2.ClusterRoleBinding, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.ClusterRoleBinding{}
	if err := a.clusterRoleBindingC.List(ctx, &slice, pred); err != nil {
		return nil, err
	}
	return slice, nil
}

// FetchClusterRoleBinding fetches an role binding resource from the backend, if authorized.
func (a *RBACClient) FetchClusterRoleBinding(ctx context.Context, name string) (*corev2.ClusterRoleBinding, error) {
	var rb corev2.ClusterRoleBinding
	if err := a.clusterRoleBindingC.Get(ctx, name, &rb); err != nil {
		return nil, err
	}
	return &rb, nil
}

// CreateClusterRoleBinding creates an role binding resource, if authorized.
func (a *RBACClient) CreateClusterRoleBinding(ctx context.Context, rb *corev2.ClusterRoleBinding) error {
	if err := a.clusterRoleBindingC.Create(ctx, rb); err != nil {
		return err
	}
	return nil
}

// UpdateClusterRoleBinding updates an role binding resource, if authorized.
func (a *RBACClient) UpdateClusterRoleBinding(ctx context.Context, rb *corev2.ClusterRoleBinding) error {
	if err := a.clusterRoleBindingC.Update(ctx, rb); err != nil {
		return err
	}
	return nil
}

// ListClusterRoles fetches a list of role resources, if authorized.
func (a *RBACClient) ListClusterRoles(ctx context.Context) ([]*corev2.ClusterRole, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.ClusterRole{}
	if err := a.clusterRoleC.List(ctx, &slice, pred); err != nil {
		return nil, err
	}
	return slice, nil
}

// FetchClusterRole fetches an role resource from the backend, if authorized.
func (a *RBACClient) FetchClusterRole(ctx context.Context, name string) (*corev2.ClusterRole, error) {
	var rb corev2.ClusterRole
	if err := a.clusterRoleC.Get(ctx, name, &rb); err != nil {
		return nil, err
	}
	return &rb, nil
}

// CreateClusterRole creates an role resource, if authorized.
func (a *RBACClient) CreateClusterRole(ctx context.Context, rb *corev2.ClusterRole) error {
	if err := a.clusterRoleC.Create(ctx, rb); err != nil {
		return err
	}
	return nil
}

// UpdateClusterRole updates an role resource, if authorized.
func (a *RBACClient) UpdateClusterRole(ctx context.Context, rb *corev2.ClusterRole) error {
	if err := a.clusterRoleC.Update(ctx, rb); err != nil {
		return err
	}
	return nil
}

// DeleteRole deletes a role resource, if authorized.
func (a *RBACClient) DeleteRole(ctx context.Context, name string) error {
	if err := a.roleC.Delete(ctx, name); err != nil {
		return err
	}
	return nil
}

// DeleteRoleBinding deletes a rolebinding resource, if authorized.
func (a *RBACClient) DeleteRoleBinding(ctx context.Context, name string) error {
	if err := a.roleBindingC.Delete(ctx, name); err != nil {
		return err
	}
	return nil
}

// DeleteClusterRole deletes a role resource, if authorized.
func (a *RBACClient) DeleteClusterRole(ctx context.Context, name string) error {
	if err := a.clusterRoleC.Delete(ctx, name); err != nil {
		return err
	}
	return nil
}

// DeleteClusterRoleBinding deletes a rolebinding resource, if authorized.
func (a *RBACClient) DeleteClusterRoleBinding(ctx context.Context, name string) error {
	if err := a.clusterRoleBindingC.Delete(ctx, name); err != nil {
		return err
	}
	return nil
}
