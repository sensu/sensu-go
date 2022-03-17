package util_api

import (
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
)

func TestInputToTypeMeta(t *testing.T) {
	tests := []struct {
		name string
		args *schema.TypeMetaInput
		want *corev2.TypeMeta
	}{{
		name: "nil",
		args: nil,
		want: nil,
	}, {
		name: "filled",
		args: &schema.TypeMetaInput{
			Type:       "CheckConfig",
			ApiVersion: "core/v2",
		},
		want: &corev2.TypeMeta{
			Type:       "CheckConfig",
			APIVersion: "core/v2",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InputToTypeMeta(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InputToTypeMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToObjectMeta(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
		want corev2.ObjectMeta
	}{{
		name: "core/v2",
		args: &corev2.Asset{
			ObjectMeta: corev2.ObjectMeta{Name: "name", Namespace: "default"},
		},
		want: corev2.ObjectMeta{Name: "name", Namespace: "default"},
	}, {
		name: "core/v2",
		args: &corev3.EntityConfig{
			Metadata: &corev2.ObjectMeta{Name: "name", Namespace: "default"},
		},
		want: corev2.ObjectMeta{Name: "name", Namespace: "default"},
	}, {
		name: "*meta",
		args: &corev2.ObjectMeta{Name: "name", Namespace: "default"},
		want: corev2.ObjectMeta{Name: "name", Namespace: "default"},
	}, {
		name: "meta",
		args: corev2.ObjectMeta{Name: "name", Namespace: "default"},
		want: corev2.ObjectMeta{Name: "name", Namespace: "default"},
	}, {
		name: "nil",
		args: nil,
		want: corev2.ObjectMeta{},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToObjectMeta(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToObjectMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}
