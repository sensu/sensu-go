package registry

// automatically generated file, do not edit!

import (
	"fmt"

	"github.com/sensu/sensu-go/internal/apis/core"
	"github.com/sensu/sensu-go/internal/apis/meta"
	"github.com/sensu/sensu-go/internal/apis/rbac"
)

type registry map[meta.TypeMeta]meta.GroupVersionKind

var typeRegistry = registry{
	meta.TypeMeta{Kind: "Namespace", APIVersion: "core"}:          core.Namespace{},
	meta.TypeMeta{Kind: "Role", APIVersion: "rbac"}:               rbac.Role{},
	meta.TypeMeta{Kind: "ClusterRole", APIVersion: "rbac"}:        rbac.ClusterRole{},
	meta.TypeMeta{Kind: "ClusterRoleBinding", APIVersion: "rbac"}: rbac.ClusterRoleBinding{},
	meta.TypeMeta{Kind: "RoleBinding", APIVersion: "rbac"}:        rbac.RoleBinding{},
	meta.TypeMeta{Kind: "Subject", APIVersion: "rbac"}:            rbac.Subject{},
}

// Resolve returns a zero-valued meta.GroupVersionKind, given a meta.TypeMeta.
// If the type does not exist, then an error will be returned.
func Resolve(mt meta.TypeMeta) (meta.GroupVersionKind, error) {
	t, ok := typeRegistry[mt]
	if !ok {
		return nil, fmt.Errorf("type could not be found: %v", mt)
	}
	return t, nil
}
