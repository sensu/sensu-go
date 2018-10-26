package registry

// automatically generated file, do not edit!

import (
	"fmt"

	"github.com/sensu/sensu-go/apis/meta/v1alpha1"
	"github.com/sensu/sensu-go/internal/apis/core"
	"github.com/sensu/sensu-go/internal/apis/meta"
	"github.com/sensu/sensu-go/internal/apis/rbac"
)

type registry map[meta.TypeMeta]interface{}

var typeRegistry = registry{
	meta.TypeMeta{APIVersion: "core", Kind: "Namespace"}:          core.Namespace{},
	meta.TypeMeta{APIVersion: "core", Kind: "namespace"}:          core.Namespace{},
	meta.TypeMeta{APIVersion: "meta", Kind: "ObjectMeta"}:         meta.ObjectMeta{},
	meta.TypeMeta{APIVersion: "meta", Kind: "objectmeta"}:         meta.ObjectMeta{},
	meta.TypeMeta{APIVersion: "meta", Kind: "TypeMeta"}:           meta.TypeMeta{},
	meta.TypeMeta{APIVersion: "meta", Kind: "typemeta"}:           meta.TypeMeta{},
	meta.TypeMeta{APIVersion: "rbac", Kind: "ClusterRole"}:        rbac.ClusterRole{},
	meta.TypeMeta{APIVersion: "rbac", Kind: "clusterrole"}:        rbac.ClusterRole{},
	meta.TypeMeta{APIVersion: "rbac", Kind: "ClusterRoleBinding"}: rbac.ClusterRoleBinding{},
	meta.TypeMeta{APIVersion: "rbac", Kind: "clusterrolebinding"}: rbac.ClusterRoleBinding{},
	meta.TypeMeta{APIVersion: "rbac", Kind: "Role"}:               rbac.Role{},
	meta.TypeMeta{APIVersion: "rbac", Kind: "role"}:               rbac.Role{},
	meta.TypeMeta{APIVersion: "rbac", Kind: "RoleBinding"}:        rbac.RoleBinding{},
	meta.TypeMeta{APIVersion: "rbac", Kind: "rolebinding"}:        rbac.RoleBinding{},
	meta.TypeMeta{APIVersion: "rbac", Kind: "Subject"}:            rbac.Subject{},
	meta.TypeMeta{APIVersion: "rbac", Kind: "subject"}:            rbac.Subject{},
	meta.TypeMeta{APIVersion: "v1alpha1", Kind: "ObjectMeta"}:     v1alpha1.ObjectMeta{},
	meta.TypeMeta{APIVersion: "v1alpha1", Kind: "objectmeta"}:     v1alpha1.ObjectMeta{},
	meta.TypeMeta{APIVersion: "v1alpha1", Kind: "TypeMeta"}:       v1alpha1.TypeMeta{},
	meta.TypeMeta{APIVersion: "v1alpha1", Kind: "typemeta"}:       v1alpha1.TypeMeta{},
}

// Resolve returns a zero-valued meta.GroupVersionKind, given a meta.TypeMeta.
// If the type does not exist, then an error will be returned.
func Resolve(mt meta.TypeMeta) (interface{}, error) {
	t, ok := typeRegistry[mt]
	if !ok {
		return nil, fmt.Errorf("type could not be found: %v", mt)
	}
	return t, nil
}
