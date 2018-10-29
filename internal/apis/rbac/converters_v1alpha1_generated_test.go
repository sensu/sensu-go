package rbac

import (
	"reflect"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/sensu/sensu-go/apis/rbac/v1alpha1"
)

func Test_Convert_Convert_rbac_ClusterRole_To_v1alpha1_ClusterRole_And_Convert_v1alpha1_ClusterRole_To_rbac_ClusterRole(t *testing.T) {
	var v1, v2 ClusterRole
	var v3 v1alpha1.ClusterRole
	fuzzer := fuzz.New().NilChance(0)
	fuzzer.Fuzz(&v1)
	if err := Convert_rbac_ClusterRole_To_v1alpha1_ClusterRole(&v3, &v1); err != nil {
		t.Fatal(err)
	}
	if err := Convert_v1alpha1_ClusterRole_To_rbac_ClusterRole(&v2, &v3); err != nil {
		t.Fatal(err)
	}

	// Set the APIVersion so we can do a DeepEqual
	v1.APIVersion = "v1alpha1"
	if !reflect.DeepEqual(v1, v2) {
		t.Fatal("values not equal")
	}
}

func Test_Convert_Convert_rbac_ClusterRoleBinding_To_v1alpha1_ClusterRoleBinding_And_Convert_v1alpha1_ClusterRoleBinding_To_rbac_ClusterRoleBinding(t *testing.T) {
	var v1, v2 ClusterRoleBinding
	var v3 v1alpha1.ClusterRoleBinding
	fuzzer := fuzz.New().NilChance(0)
	fuzzer.Fuzz(&v1)
	if err := Convert_rbac_ClusterRoleBinding_To_v1alpha1_ClusterRoleBinding(&v3, &v1); err != nil {
		t.Fatal(err)
	}
	if err := Convert_v1alpha1_ClusterRoleBinding_To_rbac_ClusterRoleBinding(&v2, &v3); err != nil {
		t.Fatal(err)
	}

	// Set the APIVersion so we can do a DeepEqual
	v1.APIVersion = "v1alpha1"
	if !reflect.DeepEqual(v1, v2) {
		t.Fatal("values not equal")
	}
}

func Test_Convert_Convert_rbac_Role_To_v1alpha1_Role_And_Convert_v1alpha1_Role_To_rbac_Role(t *testing.T) {
	var v1, v2 Role
	var v3 v1alpha1.Role
	fuzzer := fuzz.New().NilChance(0)
	fuzzer.Fuzz(&v1)
	if err := Convert_rbac_Role_To_v1alpha1_Role(&v3, &v1); err != nil {
		t.Fatal(err)
	}
	if err := Convert_v1alpha1_Role_To_rbac_Role(&v2, &v3); err != nil {
		t.Fatal(err)
	}

	// Set the APIVersion so we can do a DeepEqual
	v1.APIVersion = "v1alpha1"
	if !reflect.DeepEqual(v1, v2) {
		t.Fatal("values not equal")
	}
}

func Test_Convert_Convert_rbac_RoleBinding_To_v1alpha1_RoleBinding_And_Convert_v1alpha1_RoleBinding_To_rbac_RoleBinding(t *testing.T) {
	var v1, v2 RoleBinding
	var v3 v1alpha1.RoleBinding
	fuzzer := fuzz.New().NilChance(0)
	fuzzer.Fuzz(&v1)
	if err := Convert_rbac_RoleBinding_To_v1alpha1_RoleBinding(&v3, &v1); err != nil {
		t.Fatal(err)
	}
	if err := Convert_v1alpha1_RoleBinding_To_rbac_RoleBinding(&v2, &v3); err != nil {
		t.Fatal(err)
	}

	// Set the APIVersion so we can do a DeepEqual
	v1.APIVersion = "v1alpha1"
	if !reflect.DeepEqual(v1, v2) {
		t.Fatal("values not equal")
	}
}

func Test_Convert_Convert_rbac_Subject_To_v1alpha1_Subject_And_Convert_v1alpha1_Subject_To_rbac_Subject(t *testing.T) {
	var v1, v2 Subject
	var v3 v1alpha1.Subject
	fuzzer := fuzz.New().NilChance(0)
	fuzzer.Fuzz(&v1)
	if err := Convert_rbac_Subject_To_v1alpha1_Subject(&v3, &v1); err != nil {
		t.Fatal(err)
	}
	if err := Convert_v1alpha1_Subject_To_rbac_Subject(&v2, &v3); err != nil {
		t.Fatal(err)
	}

	// Set the APIVersion so we can do a DeepEqual
	v1.APIVersion = "v1alpha1"
	if !reflect.DeepEqual(v1, v2) {
		t.Fatal("values not equal")
	}
}
