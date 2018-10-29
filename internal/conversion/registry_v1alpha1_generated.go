package conversion

import (
	"github.com/sensu/sensu-go/internal/apis/rbac"
)

func init() {
	registry[key{
		SourceAPIVersion: "rbac",
		DestAPIVersion:   "rbac/v1alpha1",
		Kind:             "ClusterRole",
	}] = rbac.Convert_rbac_ClusterRole_To_v1alpha1_ClusterRole

	registry[key{
		SourceAPIVersion: "rbac/v1alpha1",
		DestAPIVersion:   "rbac",
		Kind:             "ClusterRole",
	}] = rbac.Convert_v1alpha1_ClusterRole_To_rbac_ClusterRole
}

func init() {
	registry[key{
		SourceAPIVersion: "rbac",
		DestAPIVersion:   "rbac/v1alpha1",
		Kind:             "ClusterRoleBinding",
	}] = rbac.Convert_rbac_ClusterRoleBinding_To_v1alpha1_ClusterRoleBinding

	registry[key{
		SourceAPIVersion: "rbac/v1alpha1",
		DestAPIVersion:   "rbac",
		Kind:             "ClusterRoleBinding",
	}] = rbac.Convert_v1alpha1_ClusterRoleBinding_To_rbac_ClusterRoleBinding
}

func init() {
	registry[key{
		SourceAPIVersion: "rbac",
		DestAPIVersion:   "rbac/v1alpha1",
		Kind:             "Role",
	}] = rbac.Convert_rbac_Role_To_v1alpha1_Role

	registry[key{
		SourceAPIVersion: "rbac/v1alpha1",
		DestAPIVersion:   "rbac",
		Kind:             "Role",
	}] = rbac.Convert_v1alpha1_Role_To_rbac_Role
}

func init() {
	registry[key{
		SourceAPIVersion: "rbac",
		DestAPIVersion:   "rbac/v1alpha1",
		Kind:             "RoleBinding",
	}] = rbac.Convert_rbac_RoleBinding_To_v1alpha1_RoleBinding

	registry[key{
		SourceAPIVersion: "rbac/v1alpha1",
		DestAPIVersion:   "rbac",
		Kind:             "RoleBinding",
	}] = rbac.Convert_v1alpha1_RoleBinding_To_rbac_RoleBinding
}

func init() {
	registry[key{
		SourceAPIVersion: "rbac",
		DestAPIVersion:   "rbac/v1alpha1",
		Kind:             "Subject",
	}] = rbac.Convert_rbac_Subject_To_v1alpha1_Subject

	registry[key{
		SourceAPIVersion: "rbac/v1alpha1",
		DestAPIVersion:   "rbac",
		Kind:             "Subject",
	}] = rbac.Convert_v1alpha1_Subject_To_rbac_Subject
}
