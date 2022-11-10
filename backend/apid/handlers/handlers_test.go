package handlers

import (
	"encoding/json"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/testing/fixture"
)

func TestCheckMeta(t *testing.T) {
	tests := []struct {
		name     string
		resource corev2.Resource
		vars     map[string]string
		wantErr  bool
	}{
		{
			name:     "namespaces mismatch",
			resource: &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Namespace: "foo"}},
			vars:     map[string]string{"namespace": "bar"},
			wantErr:  true,
		},
		{
			name:     "name mismatch",
			resource: &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "baz"}},
			vars:     map[string]string{"id": "qux"},
			wantErr:  true,
		},
		{
			name:     "valid objectmeta",
			resource: &fixture.Resource{ObjectMeta: corev2.ObjectMeta{Name: "baz", Namespace: "foo"}},
			vars:     map[string]string{"namespace": "foo", "id": "baz"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckMeta(tt.resource, tt.vars, "id"); (err != nil) != tt.wantErr {
				t.Errorf("checkMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func marshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	bytes, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}
