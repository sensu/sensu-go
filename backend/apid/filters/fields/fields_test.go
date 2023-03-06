package fields

import (
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/types"
)

type foo struct{}

var mockSelector = func(set map[string]string) bool {
	if name, ok := set["entity.name"]; ok && name == "foo" {
		return true
	}
	if name, ok := set["check.name"]; ok && name == "test-1" {
		return true
	}
	return false
}

func TestFilter(t *testing.T) {
	resource1 := corev2.FixtureEntity("foo")
	resource2 := corev2.FixtureEntity("bar")
	resource3 := types.WrapResource(corev2.FixtureCheckConfig("test-1"))
	resource4 := types.WrapResource(corev2.FixtureCheckConfig("test-2"))

	tests := []struct {
		name       string
		resources  interface{}
		fieldsFunc FieldsFunc
		want       interface{}
	}{
		{
			name:       "resources",
			resources:  []corev3.Resource{resource1, resource2},
			fieldsFunc: corev3.EntityFields,
			want:       []corev3.Resource{resource1},
		},
		{
			name:       "wrapped resources",
			resources:  []*types.Wrapper{&resource3, &resource4},
			fieldsFunc: corev3.CheckConfigFields,
			want:       []*types.Wrapper{&resource3},
		},
		{
			name:      "unknown type",
			resources: []*foo{{}},
			want:      []*foo{{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filter(tt.resources, mockSelector, tt.fieldsFunc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
