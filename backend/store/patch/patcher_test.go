package patch

import (
	"reflect"
	"testing"
)

func TestMerge_Patch(t *testing.T) {
	tests := []struct {
		name     string
		original []byte
		patch    []byte
		want     []byte
		wantErr  bool
	}{
		{
			name:     "simple merge patch",
			original: []byte(`{"name":"foo","namespace":"default"}`),
			patch:    []byte(`{"namespace":"prod"}`),
			want:     []byte(`{"name":"foo","namespace":"prod"}`),
		},
		{
			name:     "lists are replaced",
			original: []byte(`{"labels":["foo","bar"],"name":"foo"}`),
			patch:    []byte(`{"labels":["baz","qux"]}`),
			want:     []byte(`{"labels":["baz","qux"],"name":"foo"}`),
		},
		{
			name:     "nested fields are replaced",
			original: []byte(`{"command":"echo","metadata":{"name":"foo"}}`),
			patch:    []byte(`{"metadata":{"name":"bar"}}`),
			want:     []byte(`{"command":"echo","metadata":{"name":"bar"}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merge{MergePatch: tt.patch}
			got, err := m.Patch(tt.original)
			if (err != nil) != tt.wantErr {
				t.Errorf("Merge.Patch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge.Patch() = %s, want %s", string(got), string(tt.want))
			}
		})
	}
}
