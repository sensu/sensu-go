package types

import (
	"fmt"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	v2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestWrapper_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []byte
		wantErr bool
		want    Wrapper
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
			want: Wrapper{
				TypeMeta: corev2.TypeMeta{Type: "Namespace", APIVersion: "core/v2"},
				Value:    &corev2.Namespace{Name: "foo"},
			},
		},
		{
			name:  "inner and outer ObjectMeta are filled",
			bytes: []byte(`{"type": "CheckConfig", "metadata": {"name": "foo", "namespace": "dev", "labels": {"region": "us-west-2"}, "annotations": {"managed-by": "ops"}}, "spec": {"command": "echo"}}`),
			want: Wrapper{
				TypeMeta: corev2.TypeMeta{Type: "CheckConfig", APIVersion: "core/v2"},
				ObjectMeta: v2.ObjectMeta{
					Name:        "foo",
					Namespace:   "dev",
					Labels:      map[string]string{"region": "us-west-2"},
					Annotations: map[string]string{"managed-by": "ops"},
				},
				Value: &corev2.CheckConfig{
					ObjectMeta: v2.ObjectMeta{
						Name:        "foo",
						Namespace:   "dev",
						Labels:      map[string]string{"region": "us-west-2"},
						Annotations: map[string]string{"managed-by": "ops"},
					},
					Command: "echo",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Wrapper{}
			err := w.UnmarshalJSON(tt.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Wrapper.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && !reflect.DeepEqual(w, &tt.want) {
				t.Errorf("Wrapper.UnmarshalJSON() = %#v, want %#v", *w, tt.want)
			}
		})
	}
}

func TestWrapResourceObjectMeta(t *testing.T) {
	check := FixtureCheck("foo")
	check.Labels["asdf"] = "asdf"

	wrapped := WrapResource(check)
	if !reflect.DeepEqual(wrapped.ObjectMeta, check.ObjectMeta) {
		t.Fatal("objectmeta not equal")
	}
}

func TestResolveType(t *testing.T) {
	testCases := []struct {
		ApiVersion string
		Type       string
		ExpRet     interface{}
		ExpErr     bool
	}{
		{
			ApiVersion: "core/v2",
			Type:       "asset",
			ExpRet:     &v2.Asset{},
			ExpErr:     false,
		},
		{
			ApiVersion: "non/existence",
			Type:       "null",
			ExpRet:     nil,
			ExpErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/%s", tc.ApiVersion, tc.Type), func(t *testing.T) {
			r, err := ResolveType(tc.ApiVersion, tc.Type)
			if !reflect.DeepEqual(r, tc.ExpRet) {
				t.Fatal("unexpected type")
			}
			if err != nil && !tc.ExpErr {
				t.Fatal(err)
			}
			if err == nil && tc.ExpErr {
				t.Fatal("expected an error")
			}
		})
	}
}
