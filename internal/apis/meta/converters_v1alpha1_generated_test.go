package meta

import (
	"reflect"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/sensu/sensu-go/apis/meta/v1alpha1"
)

func Test_Convert_Convert_meta_ObjectMeta_To_v1alpha1_ObjectMeta_And_Convert_v1alpha1_ObjectMeta_To_meta_ObjectMeta(t *testing.T) {
	var v1, v2 ObjectMeta
	var v3 v1alpha1.ObjectMeta
	fuzzer := fuzz.New().NilChance(0)
	fuzzer.Fuzz(&v1)
	if err := Convert_meta_ObjectMeta_To_v1alpha1_ObjectMeta(&v3, &v1); err != nil {
		t.Fatal(err)
	}
	if err := Convert_v1alpha1_ObjectMeta_To_meta_ObjectMeta(&v2, &v3); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v1, v2) {
		t.Fatal("values not equal")
	}
}

func Test_Convert_Convert_meta_TypeMeta_To_v1alpha1_TypeMeta_And_Convert_v1alpha1_TypeMeta_To_meta_TypeMeta(t *testing.T) {
	var v1, v2 TypeMeta
	var v3 v1alpha1.TypeMeta
	fuzzer := fuzz.New().NilChance(0)
	fuzzer.Fuzz(&v1)
	if err := Convert_meta_TypeMeta_To_v1alpha1_TypeMeta(&v3, &v1); err != nil {
		t.Fatal(err)
	}
	if err := Convert_v1alpha1_TypeMeta_To_meta_TypeMeta(&v2, &v3); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v1, v2) {
		t.Fatal("values not equal")
	}
}
