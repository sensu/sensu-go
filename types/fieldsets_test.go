package types

import (
	"testing"
)

func TestLookupFieldsetFn(t *testing.T) {
	type args struct {
		apiVersion string
		typename   string
	}
	tests := []struct {
		name   string
		args   args
		wantOk bool
	}{
		{
			name: "not found",
			args: args{
				apiVersion: "core/v2",
				typename:   "unknown",
			},
			wantOk: false,
		},
		{
			name: "want Event",
			args: args{
				apiVersion: "core/v2",
				typename:   "Event",
			},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := LookupFieldsetFn(tt.args.apiVersion, tt.args.typename)
			if got != tt.wantOk {
				t.Errorf("LookupFieldsetFn() got1 = %v, want %v", got, tt.wantOk)
			}
		})
	}
}
