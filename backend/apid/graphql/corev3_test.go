package graphql

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	util_api "github.com/sensu/sensu-go/backend/apid/graphql/util/api"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/mock"
)

func Test_corev3_ID(t *testing.T) {
	tests := []struct {
		name     string
		resolver interface {
			ID(p graphql.ResolveParams) (string, error)
		}
		in      interface{}
		want    string
		wantErr bool
	}{
		{
			name:     "corev3EntityConfigExtImpl",
			resolver: &corev3EntityConfigExtImpl{},
			in:       corev3.FixtureEntityConfig("test"),
			want:     "srn:core/v3.EntityConfig:default:test",
			wantErr:  false,
		},
		{
			name:     "corev3EntityStateExtImpl",
			resolver: &corev3EntityStateExtImpl{},
			in:       corev3.FixtureEntityState("test"),
			want:     "srn:core/v3.EntityState:default:test",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T/%s", tt.resolver, tt.name), func(t *testing.T) {
			params := graphql.ResolveParams{Context: context.Background(), Source: tt.in}
			got, err := tt.resolver.ID(params)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s.ID() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("%s.ID() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_corev3_Metadata(t *testing.T) {
	tests := []struct {
		name     string
		resolver interface {
			Metadata(p graphql.ResolveParams) (interface{}, error)
		}
		in      interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:     "corev3EntityConfigExtImpl",
			resolver: &corev3EntityConfigExtImpl{},
			in: func() *corev3.EntityConfig {
				obj := corev3.FixtureEntityConfig("name")
				obj.Metadata.Labels["my_label"] = "test"
				obj.Metadata.Labels["password"] = "test"
				return obj
			}(),
			want: func() *corev2.ObjectMeta {
				obj := corev3.FixtureEntityConfig("name")
				obj.Metadata.Labels["my_label"] = "test"
				obj.Metadata.Labels["password"] = corev2.Redacted
				return obj.Metadata
			}(),
			wantErr: false,
		},
		{
			name:     "corev3EntityStateExtImpl",
			resolver: &corev3EntityStateExtImpl{},
			in:       corev3.FixtureEntityState("name"),
			want: func() *corev2.ObjectMeta {
				obj := corev3.FixtureEntityState("name")
				return obj.Metadata
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T/%s", tt.resolver, tt.name), func(t *testing.T) {
			got, err := tt.resolver.Metadata(graphql.ResolveParams{Context: context.Background(), Source: tt.in})
			if (err != nil) != tt.wantErr {
				t.Errorf("%s.Metadata() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s.Metadata() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_corev3_ToJSON(t *testing.T) {
	tests := []struct {
		name     string
		resolver interface {
			ToJSON(p graphql.ResolveParams) (interface{}, error)
		}
		in      interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:     "corev3EntityConfigExtImpl",
			resolver: &corev3EntityConfigExtImpl{},
			in:       corev3.FixtureEntityConfig("name"),
			want:     util_api.WrapResource(corev3.FixtureEntityConfig("name")),
			wantErr:  false,
		},
		{
			name:     "corev3EntityStateExtImpl",
			resolver: &corev3EntityStateExtImpl{},
			in:       corev3.FixtureEntityState("name"),
			want:     util_api.WrapResource(corev3.FixtureEntityState("name")),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T/%s", tt.resolver, tt.name), func(t *testing.T) {
			got, err := tt.resolver.ToJSON(graphql.ResolveParams{Context: context.Background(), Source: tt.in})
			if (err != nil) != tt.wantErr {
				t.Errorf("%s.ToJSON() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s.ToJSON() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_corev3_IsTypeOf(t *testing.T) {
	tests := []struct {
		name     string
		resolver interface {
			IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool
		}
		in   interface{}
		want bool
	}{
		{
			name:     "entity_config/match",
			resolver: &corev3EntityConfigImpl{},
			in:       corev3.FixtureEntityConfig("name"),
			want:     true,
		},
		{
			name:     "entity_config/no match",
			resolver: &corev3EntityConfigImpl{},
			in:       corev2.FixtureEntity("name"),
			want:     false,
		},
		{
			name:     "entity_state/match",
			resolver: &corev3EntityStateImpl{},
			in:       corev3.FixtureEntityState("name"),
			want:     true,
		},
		{
			name:     "entity_state/no match",
			resolver: &corev3EntityStateImpl{},
			in:       corev2.FixtureEntity("name"),
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T/%s", tt.resolver, tt.name), func(t *testing.T) {
			got := tt.resolver.IsTypeOf(tt.in, graphql.IsTypeOfParams{Context: context.Background()})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%T.ToJSON() = %v, want %v", tt.resolver, got, tt.want)
			}
		})
	}
}

func Test_corev3EntityConfigExtImpl_State(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockGenericClient)
		source  interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "success",
			setup: func(client *MockGenericClient) {
				client.On("SetTypeMeta", mock.Anything).Return(nil)
				client.On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev3.V2ResourceProxy)
					*arg = corev3.V2ResourceProxy{Resource: corev3.FixtureEntityConfig("name")}
				}).Return(nil).Once()
			},
			source:  corev3.FixtureEntityConfig("name"),
			want:    corev3.FixtureEntityConfig("name"),
			wantErr: false,
		},
		{
			name: "not found",
			setup: func(client *MockGenericClient) {
				client.On("SetTypeMeta", mock.Anything).Return(nil)
				client.On("Get", mock.Anything, "name", mock.Anything).Return(&store.ErrNotFound{}).Once()
			},
			source:  corev3.FixtureEntityConfig("name"),
			want:    nil,
			wantErr: false,
		},
		{
			name: "bad meta",
			setup: func(client *MockGenericClient) {
				client.On("SetTypeMeta", mock.Anything).Return(errors.New("unlikely"))
				client.On("Get", mock.Anything, "name", mock.Anything).Return(&store.ErrNotFound{}).Once()
			},
			source:  corev3.FixtureEntityConfig("name"),
			want:    nil,
			wantErr: true,
		},
		{
			name: "upstream err",
			setup: func(client *MockGenericClient) {
				client.On("SetTypeMeta", mock.Anything).Return(nil)
				client.On("Get", mock.Anything, "name", mock.Anything).Return(errors.New("unlikely")).Once()
			},
			source:  corev3.FixtureEntityConfig("name"),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(MockGenericClient)
			tt.setup(client)

			impl := &corev3EntityConfigExtImpl{client: client}
			got, err := impl.State(graphql.ResolveParams{
				Context: context.Background(),
				Source:  tt.source,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("corev3EntityConfigExtImpl.State() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("corev3EntityConfigExtImpl.State() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_corev3EntityStateExtImpl_State(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockGenericClient)
		source  interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "success",
			setup: func(client *MockGenericClient) {
				client.On("SetTypeMeta", mock.Anything).Return(nil)
				client.On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev3.V2ResourceProxy)
					*arg = corev3.V2ResourceProxy{Resource: corev3.FixtureEntityState("name")}
				}).Return(nil).Once()
			},
			source:  corev3.FixtureEntityState("name"),
			want:    corev3.FixtureEntityState("name"),
			wantErr: false,
		},
		{
			name: "not found",
			setup: func(client *MockGenericClient) {
				client.On("SetTypeMeta", mock.Anything).Return(nil)
				client.On("Get", mock.Anything, "name", mock.Anything).Return(&store.ErrNotFound{}).Once()
			},
			source:  corev3.FixtureEntityState("name"),
			want:    nil,
			wantErr: false,
		},
		{
			name: "bad meta",
			setup: func(client *MockGenericClient) {
				client.On("SetTypeMeta", mock.Anything).Return(errors.New("unlikely"))
				client.On("Get", mock.Anything, "name", mock.Anything).Return(&store.ErrNotFound{}).Once()
			},
			source:  corev3.FixtureEntityState("name"),
			want:    nil,
			wantErr: true,
		},
		{
			name: "upstream err",
			setup: func(client *MockGenericClient) {
				client.On("SetTypeMeta", mock.Anything).Return(nil)
				client.On("Get", mock.Anything, "name", mock.Anything).Return(errors.New("unlikely")).Once()
			},
			source:  corev3.FixtureEntityState("name"),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(MockGenericClient)
			tt.setup(client)

			impl := &corev3EntityStateExtImpl{client: client}
			got, err := impl.Config(graphql.ResolveParams{
				Context: context.Background(),
				Source:  tt.source,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("corev3EntityStateExtImpl.Config() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("corev3EntityStateExtImpl.Config() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_corev3EntityStateExtImpl_ToCoreV2Entity(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockEntityClient)
		source  interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "success",
			setup: func(client *MockEntityClient) {
				client.On("FetchEntity", mock.Anything, "name").
					Return(corev2.FixtureEntity("name"), nil).
					Once()
			},
			source:  corev3.FixtureEntityState("name"),
			want:    corev2.FixtureEntity("name"),
			wantErr: false,
		},
		{
			name: "upstream err",
			setup: func(client *MockEntityClient) {
				client.On("FetchEntity", mock.Anything, "name").
					Return(corev2.FixtureEntity("name"), errors.New("unlikely")).
					Once()
			},
			source:  corev3.FixtureEntityState("name"),
			want:    corev2.FixtureEntity("name"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(MockEntityClient)
			tt.setup(client)

			impl := &corev3EntityStateExtImpl{entityClient: client}
			got, err := impl.ToCoreV2Entity(graphql.ResolveParams{
				Context: context.Background(),
				Source:  tt.source,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("corev3EntityStateExtImpl.ToCoreV2Entity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("corev3EntityStateExtImpl.ToCoreV2Entity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_corev3EntityConfigExtImpl_ToCoreV2Entity(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockEntityClient)
		source  interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "success",
			setup: func(client *MockEntityClient) {
				client.On("FetchEntity", mock.Anything, "name").
					Return(corev2.FixtureEntity("name"), nil).
					Once()
			},
			source:  corev3.FixtureEntityConfig("name"),
			want:    corev2.FixtureEntity("name"),
			wantErr: false,
		},
		{
			name: "upstream err",
			setup: func(client *MockEntityClient) {
				client.On("FetchEntity", mock.Anything, "name").
					Return(corev2.FixtureEntity("name"), errors.New("unlikely")).
					Once()
			},
			source:  corev3.FixtureEntityConfig("name"),
			want:    corev2.FixtureEntity("name"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(MockEntityClient)
			tt.setup(client)

			impl := &corev3EntityConfigExtImpl{entityClient: client}
			got, err := impl.ToCoreV2Entity(graphql.ResolveParams{
				Context: context.Background(),
				Source:  tt.source,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("corev3EntityConfigExtImpl.ToCoreV2Entity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("corev3EntityConfigExtImpl.ToCoreV2Entity() = %v, want %v", got, tt.want)
			}
		})
	}
}
