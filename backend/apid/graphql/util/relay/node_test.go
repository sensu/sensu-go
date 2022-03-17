package util_relay

import (
	"context"
	"errors"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/stretchr/testify/mock"
)

type MockFetcher struct {
	mock.Mock
}

func (c *MockFetcher) SetTypeMeta(meta corev2.TypeMeta) error {
	return c.Called(meta).Error(0)
}

func (c *MockFetcher) Get(ctx context.Context, name string, val corev2.Resource) error {
	return c.Called(ctx, name, val).Error(0)
}

func TestToGID(t *testing.T) {
	type args struct {
		ctx context.Context
		r   interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "check",
			args: args{
				ctx: context.Background(),
				r:   corev2.FixtureCheckConfig("name"),
			},
			want: "srn:checks:default:name",
		},
		{
			name: "unknown",
			args: args{
				ctx: context.Background(),
				r:   struct{}{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToGID(tt.args.ctx, tt.args.r); got != tt.want {
				t.Errorf("ToGID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeNodeResolver(t *testing.T) {
	tests := []struct {
		name     string
		meta     corev2.TypeMeta
		setup    func() Fetcher
		globalid string
		want     interface{}
		wantErr  bool
	}{
		{
			name: "success",
			meta: corev2.TypeMeta{
				APIVersion: "core/v2",
				Type:       "pipeline",
			},
			globalid: "srn:corev2/pipeline:default:name",
			setup: func() Fetcher {
				fetcher := &MockFetcher{}
				fetcher.On("SetTypeMeta", mock.Anything).Return(nil).Once()
				fetcher.On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev2.Pipeline)
					*arg = *corev2.FixturePipeline("name", "default")
				}).Return(nil).Once()
				return fetcher
			},
			want:    corev2.FixturePipeline("name", "default"),
			wantErr: false,
		},
		{
			name: "bad type meta",
			meta: corev2.TypeMeta{
				APIVersion: "core/v2",
				Type:       "pipeline",
			},
			globalid: "srn:corev2/pipeline:default:name",
			setup: func() Fetcher {
				fetcher := &MockFetcher{}
				fetcher.On("SetTypeMeta", mock.Anything).Return(errors.New("test")).Once()
				fetcher.On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev2.Pipeline)
					*arg = *corev2.FixturePipeline("name", "default")
				}).Return(nil).Once()
				return fetcher
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "cannot resolve type",
			meta: corev2.TypeMeta{
				APIVersion: "tm/v1",
				Type:       "spoof",
			},
			globalid: "srn:corev2/pipeline:default:name",
			setup: func() Fetcher {
				fetcher := &MockFetcher{}
				fetcher.On("SetTypeMeta", mock.Anything).Return(nil).Once()
				fetcher.On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev2.Pipeline)
					*arg = *corev2.FixturePipeline("name", "default")
				}).Return(nil).Once()
				return fetcher
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := MakeNodeResolver(tt.setup(), tt.meta)
			if resolver == nil {
				t.Fatalf("MakeNodeResolver() = nil")
			}

			g, err := globalid.Decode(tt.globalid)
			if err != nil {
				t.Fatalf("unable to decode globalid: %s", err)
			}

			got, err := resolver(relay.NodeResolverParams{
				Context:      context.Background(),
				IDComponents: g,
			})
			if err != nil && !tt.wantErr {
				t.Errorf("didn't expect err but got: %s", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expect %v, got: %v", got, err)
			}
		})
	}
}
