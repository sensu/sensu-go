package stringutil

import (
	"reflect"
	"testing"
)

func TestMergeMapWithPrefix(t *testing.T) {
	type args struct {
		a      map[string]string
		b      map[string]string
		prefix string
	}
	tests := []struct {
		name   string
		args   args
		expect map[string]string
	}{
		{
			name: "merge empty",
			args: args{
				a:      map[string]string{},
				b:      map[string]string{},
				prefix: "prefix.",
			},
			expect: map[string]string{},
		},
		{
			name: "empty prefix",
			args: args{
				a:      map[string]string{"a": "b"},
				b:      map[string]string{"a": "c"},
				prefix: "",
			},
			expect: map[string]string{"a": "c"},
		},
		{
			name: "with prefix",
			args: args{
				a:      map[string]string{"a": "b"},
				b:      map[string]string{"a": "b"},
				prefix: "c.",
			},
			expect: map[string]string{"a": "b", "c.a": "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MergeMapWithPrefix(tt.args.a, tt.args.b, tt.args.prefix)
			if !reflect.DeepEqual(tt.args.a, tt.expect) {
				t.Errorf("MergeMapWithPrefix() = %#v, want %#v", tt.args.a, tt.expect)
			}
		})
	}
}
