package handlers

import (
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type mockResource struct {
	corev2.ObjectMeta
}

func (r *mockResource) GetObjectMeta() corev2.ObjectMeta {
	return r.ObjectMeta
}

func (r *mockResource) StorePath() string {
	return ""
}

func (r *mockResource) URIPath() string {
	return ""
}

func (r *mockResource) Validate() error {
	return nil
}

func TestCheckMeta(t *testing.T) {
	tests := []struct {
		name     string
		resource corev2.Resource
		vars     map[string]string
		wantErr  bool
	}{
		{
			name:     "namespaces mismatch",
			resource: &mockResource{ObjectMeta: corev2.ObjectMeta{Namespace: "foo"}},
			vars:     map[string]string{"namespace": "bar"},
			wantErr:  true,
		},
		{
			name:     "name mismatch",
			resource: &mockResource{ObjectMeta: corev2.ObjectMeta{Name: "baz"}},
			vars:     map[string]string{"id": "qux"},
			wantErr:  true,
		},
		{
			name:     "valid objectmeta",
			resource: &mockResource{ObjectMeta: corev2.ObjectMeta{Name: "baz", Namespace: "foo"}},
			vars:     map[string]string{"namespace": "foo", "id": "baz"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkMeta(tt.resource, tt.vars); (err != nil) != tt.wantErr {
				t.Errorf("checkMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
