package util_api

import (
	"reflect"
	"testing"
)

func TestMakeKVPairString(t *testing.T) {
	tests := []struct {
		name string
		args map[string]string
		want []KVPairString
	}{{
		name: "nil",
		args: nil,
		want: []KVPairString{},
	}, {
		name: "map",
		args: map[string]string{
			"key": "val",
		},
		want: []KVPairString{
			{Key: "key", Val: "val"},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeKVPairString(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeKVPairString() = %v, want %v", got, tt.want)
			}
		})
	}
}
