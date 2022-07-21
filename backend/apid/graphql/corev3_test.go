package graphql

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	util_api "github.com/sensu/sensu-go/backend/apid/graphql/util/api"
	"github.com/sensu/sensu-go/graphql"
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
