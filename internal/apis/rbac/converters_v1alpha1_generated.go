package rbac

import (
	"unsafe"

	"github.com/sensu/sensu-go/apis/rbac/v1alpha1"
)

func Convert_rbac_ClusterRole_To_v1alpha1_ClusterRole(dst, src interface{}) error {
	dstp := dst.(*v1alpha1.ClusterRole)
	srcp := src.(*ClusterRole)

	*dstp = *(*v1alpha1.ClusterRole)(unsafe.Pointer(srcp))

	return nil
}

func Convert_v1alpha1_ClusterRole_To_rbac_ClusterRole(dst, src interface{}) error {
	dstp := dst.(*ClusterRole)
	srcp := src.(*v1alpha1.ClusterRole)

	*dstp = *(*ClusterRole)(unsafe.Pointer(srcp))

	return nil
}

func Convert_rbac_ClusterRoleBinding_To_v1alpha1_ClusterRoleBinding(dst, src interface{}) error {
	dstp := dst.(*v1alpha1.ClusterRoleBinding)
	srcp := src.(*ClusterRoleBinding)

	*dstp = *(*v1alpha1.ClusterRoleBinding)(unsafe.Pointer(srcp))

	return nil
}

func Convert_v1alpha1_ClusterRoleBinding_To_rbac_ClusterRoleBinding(dst, src interface{}) error {
	dstp := dst.(*ClusterRoleBinding)
	srcp := src.(*v1alpha1.ClusterRoleBinding)

	*dstp = *(*ClusterRoleBinding)(unsafe.Pointer(srcp))

	return nil
}

func Convert_rbac_Role_To_v1alpha1_Role(dst, src interface{}) error {
	dstp := dst.(*v1alpha1.Role)
	srcp := src.(*Role)

	*dstp = *(*v1alpha1.Role)(unsafe.Pointer(srcp))

	return nil
}

func Convert_v1alpha1_Role_To_rbac_Role(dst, src interface{}) error {
	dstp := dst.(*Role)
	srcp := src.(*v1alpha1.Role)

	*dstp = *(*Role)(unsafe.Pointer(srcp))

	return nil
}

func Convert_rbac_RoleBinding_To_v1alpha1_RoleBinding(dst, src interface{}) error {
	dstp := dst.(*v1alpha1.RoleBinding)
	srcp := src.(*RoleBinding)

	*dstp = *(*v1alpha1.RoleBinding)(unsafe.Pointer(srcp))

	return nil
}

func Convert_v1alpha1_RoleBinding_To_rbac_RoleBinding(dst, src interface{}) error {
	dstp := dst.(*RoleBinding)
	srcp := src.(*v1alpha1.RoleBinding)

	*dstp = *(*RoleBinding)(unsafe.Pointer(srcp))

	return nil
}

func Convert_rbac_Subject_To_v1alpha1_Subject(dst, src interface{}) error {
	dstp := dst.(*v1alpha1.Subject)
	srcp := src.(*Subject)

	*dstp = *(*v1alpha1.Subject)(unsafe.Pointer(srcp))

	return nil
}

func Convert_v1alpha1_Subject_To_rbac_Subject(dst, src interface{}) error {
	dstp := dst.(*Subject)
	srcp := src.(*v1alpha1.Subject)

	*dstp = *(*Subject)(unsafe.Pointer(srcp))

	return nil
}
