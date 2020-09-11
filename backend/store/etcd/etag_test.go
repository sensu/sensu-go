package etcd

import (
	"testing"
)

type BadStruct struct {
	Name        string
	BustedField int `hash:"string"`
}

type GoodStruct struct {
	Name string
}

func TestETag(t *testing.T) {
	tests := []struct {
		name    string
		v       interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "invalid resource",
			v:       &BadStruct{},
			wantErr: true,
		},
		{
			name: "valid resource",
			v:    &GoodStruct{Name: "foo"},
			want: `"13487218765270944197"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ETag(tt.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("ETag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ETag() = %v, want %v", got, tt.want)

			}
		})
	}
}
