package v1alpha1

import (
	"reflect"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/sensu/sensu-go/internal/apis/meta"
)

func Test_convert_ObjectMeta_To_meta_ObjectMeta(t *testing.T) {
	var v1, v2 ObjectMeta
	var v3 meta.ObjectMeta
	fuzzer := fuzz.New().NilChance(0)
	fuzzer.Fuzz(&v1)
	v1.ConvertTo(&v3)
	v2.ConvertFrom(&v3)
	if !reflect.DeepEqual(v1, v2) {
		t.Fatal("values not equal")
	}
}
func Test_convert_TypeMeta_To_meta_TypeMeta(t *testing.T) {
	var v1, v2 TypeMeta
	var v3 meta.TypeMeta
	fuzzer := fuzz.New().NilChance(0)
	fuzzer.Fuzz(&v1)
	v1.ConvertTo(&v3)
	v2.ConvertFrom(&v3)
	if !reflect.DeepEqual(v1, v2) {
		t.Fatal("values not equal")
	}
}
