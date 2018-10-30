package registry

// automatically generated file, do not edit!

import (
	"fmt"
	"reflect"

	metav1 "github.com/sensu/sensu-go/apis/meta/v1"
	rbacv1alpha1 "github.com/sensu/sensu-go/apis/rbac/v1alpha1"
	"github.com/sensu/sensu-go/internal/apis/core"
	"github.com/sensu/sensu-go/internal/apis/rbac"
)

type registry map[metav1.TypeMeta]interface{}

var typeRegistry = registry{
	metav1.TypeMeta{APIVersion: "core", Kind: "Namespace"}:                   core.Namespace{},
	metav1.TypeMeta{APIVersion: "core", Kind: "namespace"}:                   core.Namespace{},
	metav1.TypeMeta{APIVersion: "rbac", Kind: "ClusterRole"}:                 rbac.ClusterRole{},
	metav1.TypeMeta{APIVersion: "rbac", Kind: "clusterrole"}:                 rbac.ClusterRole{},
	metav1.TypeMeta{APIVersion: "rbac", Kind: "ClusterRoleBinding"}:          rbac.ClusterRoleBinding{},
	metav1.TypeMeta{APIVersion: "rbac", Kind: "clusterrolebinding"}:          rbac.ClusterRoleBinding{},
	metav1.TypeMeta{APIVersion: "rbac", Kind: "Role"}:                        rbac.Role{},
	metav1.TypeMeta{APIVersion: "rbac", Kind: "role"}:                        rbac.Role{},
	metav1.TypeMeta{APIVersion: "rbac", Kind: "RoleBinding"}:                 rbac.RoleBinding{},
	metav1.TypeMeta{APIVersion: "rbac", Kind: "rolebinding"}:                 rbac.RoleBinding{},
	metav1.TypeMeta{APIVersion: "rbac", Kind: "Subject"}:                     rbac.Subject{},
	metav1.TypeMeta{APIVersion: "rbac", Kind: "subject"}:                     rbac.Subject{},
	metav1.TypeMeta{APIVersion: "rbac/v1alpha1", Kind: "ClusterRole"}:        rbacv1alpha1.ClusterRole{},
	metav1.TypeMeta{APIVersion: "rbac/v1alpha1", Kind: "clusterrole"}:        rbacv1alpha1.ClusterRole{},
	metav1.TypeMeta{APIVersion: "rbac/v1alpha1", Kind: "ClusterRoleBinding"}: rbacv1alpha1.ClusterRoleBinding{},
	metav1.TypeMeta{APIVersion: "rbac/v1alpha1", Kind: "clusterrolebinding"}: rbacv1alpha1.ClusterRoleBinding{},
	metav1.TypeMeta{APIVersion: "rbac/v1alpha1", Kind: "Role"}:               rbacv1alpha1.Role{},
	metav1.TypeMeta{APIVersion: "rbac/v1alpha1", Kind: "role"}:               rbacv1alpha1.Role{},
	metav1.TypeMeta{APIVersion: "rbac/v1alpha1", Kind: "RoleBinding"}:        rbacv1alpha1.RoleBinding{},
	metav1.TypeMeta{APIVersion: "rbac/v1alpha1", Kind: "rolebinding"}:        rbacv1alpha1.RoleBinding{},
	metav1.TypeMeta{APIVersion: "rbac/v1alpha1", Kind: "Subject"}:            rbacv1alpha1.Subject{},
	metav1.TypeMeta{APIVersion: "rbac/v1alpha1", Kind: "subject"}:            rbacv1alpha1.Subject{},
}

func init() {
	for k, v := range typeRegistry {
		r, ok := v.(interface{ ResourceName() string })
		if ok {
			newKey := metav1.TypeMeta{
				APIVersion: k.APIVersion,
				Kind:       r.ResourceName(),
			}
			typeRegistry[newKey] = v
		}
	}
}

// Resolve returns a zero-valued sensu object, given a metav1.TypeMeta.
// If the type does not exist, then an error will be returned.
func Resolve(mt metav1.TypeMeta) (interface{}, error) {
	t, ok := typeRegistry[mt]
	if !ok {
		return nil, fmt.Errorf("type could not be found: %v", mt)
	}
	return t, nil
}

// ResolveSlice returns a zero-valued slice of sensu objects, given a
// meta.TypeMeta. If the type does not exist, then an error will be returned.
func ResolveSlice(mt metav1.TypeMeta) (interface{}, error) {
	t, err := Resolve(mt)
	if err != nil {
		return nil, err
	}
	return reflect.Indirect(reflect.New(reflect.SliceOf(reflect.TypeOf(t)))).Interface(), nil
}
