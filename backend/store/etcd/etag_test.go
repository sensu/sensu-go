package etcd

import (
	"encoding/hex"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/require"
)

func TestETag(t *testing.T) {
	type InvalidResource struct{}
	type args struct {
		r interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "invalid resource",
			args: args{
				r: InvalidResource{},
			},
			wantErr: true,
		},
		{
			name: "protobuf",
			args: args{
				corev2.FixtureEntity("entity1"),
			},
			want:    "dec0e8aceeb38b296ba1574b46150cc1bc51bacd",
			wantErr: false,
		},
		{
			name: "json",
			args: args{
				r: types.Wrapper{},
			},
			want:    "f8bcce9286d770601b0a85733bf3f32fc549ba50",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ETag(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ETag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != "" {
				want, err := hex.DecodeString(tt.want)
				require.NoError(t, err)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("ETag() = %v, want %v", got, want)
					t.Errorf("ETag() hex = %x, want %x", got, want)
				}
			}
		})
	}
}
