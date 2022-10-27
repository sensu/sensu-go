package types_test

import (
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	types "github.com/sensu/sensu-go/types"
)

func TestWrapper_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []byte
		wantErr bool
		want    types.Wrapper
	}{
		{
			name:    "not a wrapped-json struct",
			bytes:   []byte(`"foo"`),
			wantErr: true,
		},
		{
			name:    "unresolved type",
			bytes:   []byte(`{"type": "Foo"}`),
			wantErr: true,
		},
		{
			name:    "no spec object",
			bytes:   []byte(`{"type": "CheckConfig"}`),
			wantErr: true,
		},
		{
			name:    "invalid spec",
			bytes:   []byte(`{"type": "CheckConfig", "spec": "foo"}`),
			wantErr: true,
		},
		{
			name:  "namespace resource",
			bytes: []byte(`{"type": "Namespace", "spec": {"name": "foo"}}`),
			want: types.Wrapper{
				TypeMeta: corev2.TypeMeta{Type: "Namespace", APIVersion: "core/v2"},
				Value:    &corev2.Namespace{Name: "foo"},
			},
		},
		{
			name:  "inner and outer ObjectMeta are filled",
			bytes: []byte(`{"type": "CheckConfig", "metadata": {"name": "foo", "namespace": "dev", "labels": {"region": "us-west-2"}, "annotations": {"managed-by": "ops"}}, "spec": {"command": "echo"}}`),
			want: types.Wrapper{
				TypeMeta: corev2.TypeMeta{Type: "CheckConfig", APIVersion: "core/v2"},
				ObjectMeta: corev2.ObjectMeta{
					Name:        "foo",
					Namespace:   "dev",
					Labels:      map[string]string{"region": "us-west-2"},
					Annotations: map[string]string{"managed-by": "ops"},
				},
				Value: &corev2.CheckConfig{
					ObjectMeta: corev2.ObjectMeta{
						Name:        "foo",
						Namespace:   "dev",
						Labels:      map[string]string{"region": "us-west-2"},
						Annotations: map[string]string{"managed-by": "ops"},
					},
					Command: "echo",
				},
			},
		},
		{
			name:  "inner & outer ObjectMeta are filled for core/v3 resource",
			bytes: []byte(`{"type": "EntityConfig", "api_version": "core/v3", "metadata": {"name": "foo", "namespace": "dev"}, "spec": {"entity_class": "agent", "subscriptions": ["testsub"]}}`),
			want: types.Wrapper{
				TypeMeta: corev2.TypeMeta{Type: "EntityConfig", APIVersion: "core/v3"},
				ObjectMeta: corev2.ObjectMeta{
					Name:      "foo",
					Namespace: "dev",
				},
				Value: &corev3.EntityConfig{
					Metadata: &corev2.ObjectMeta{
						Name:        "foo",
						Namespace:   "dev",
						Labels:      map[string]string{},
						Annotations: map[string]string{},
					},
					EntityClass: "agent",
					Subscriptions: []string{
						"testsub",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &types.Wrapper{}
			err := w.UnmarshalJSON(tt.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Wrapper.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && !reflect.DeepEqual(w, &tt.want) {
				t.Errorf("Wrapper.UnmarshalJSON() = %#v, \nwant %#v", *w, tt.want)
			}
		})
	}
}

func TestWrapResourceObjectMeta(t *testing.T) {
	check := corev2.FixtureCheck("foo")
	check.Labels["asdf"] = "asdf"

	wrapped := types.WrapResource(check)
	if !reflect.DeepEqual(wrapped.ObjectMeta, check.ObjectMeta) {
		t.Fatal("objectmeta not equal")
	}
}
