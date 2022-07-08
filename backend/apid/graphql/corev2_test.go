package graphql

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

func Test_corev2_ID(t *testing.T) {
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
			name:     "default",
			resolver: &corev2PipelineImpl{},
			in:       corev2.FixturePipeline("test", "default"),
			want:     "srn:corev2/pipeline:default:test",
			wantErr:  false,
		},
		{
			name:     "role",
			resolver: &roleImpl{},
			in:       corev2.FixtureRole("test", "default"),
			want:     "srn:roles:default:test",
			wantErr:  false,
		},
		{
			name:     "role_binding",
			resolver: &roleBindingImpl{},
			in:       corev2.FixtureRoleBinding("test", "default"),
			want:     "srn:rolebindings:default:test",
			wantErr:  false,
		},
		{
			name:     "cluster_role",
			resolver: &clusterRoleImpl{},
			in:       corev2.FixtureClusterRole("test"),
			want:     "srn:clusterroles:test",
			wantErr:  false,
		},
		{
			name:     "cluster_role_binding",
			resolver: &clusterRoleBindingImpl{},
			in:       corev2.FixtureClusterRoleBinding("test"),
			want:     "srn:clusterrolebindings:test",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T/%s", tt.resolver, tt.name), func(t *testing.T) {
			params := graphql.ResolveParams{Context: context.Background(), Source: tt.in}
			got, err := tt.resolver.ID(params)
			if (err != nil) != tt.wantErr {
				t.Errorf("corev2PipelineImpl.ID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("corev2PipelineImpl.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_corev2_ToJSON(t *testing.T) {
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
			name:     "default",
			resolver: &corev2PipelineImpl{},
			in:       corev2.FixturePipeline("name", "default"),
			want:     types.WrapResource(corev2.FixturePipeline("name", "default")),
			wantErr:  false,
		},
		{
			name:     "default",
			resolver: &roleImpl{},
			in:       corev2.FixtureRole("name", "default"),
			want:     types.WrapResource(corev2.FixtureRole("name", "default")),
			wantErr:  false,
		},
		{
			name:     "default",
			resolver: &roleBindingImpl{},
			in:       corev2.FixtureRoleBinding("name", "default"),
			want:     types.WrapResource(corev2.FixtureRoleBinding("name", "default")),
			wantErr:  false,
		},
		{
			name:     "default",
			resolver: &clusterRoleImpl{},
			in:       corev2.FixtureClusterRole("name"),
			want:     types.WrapResource(corev2.FixtureClusterRole("name")),
			wantErr:  false,
		},
		{
			name:     "default",
			resolver: &clusterRoleBindingImpl{},
			in:       corev2.FixtureClusterRoleBinding("name"),
			want:     types.WrapResource(corev2.FixtureClusterRoleBinding("name")),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T/%s", tt.resolver, tt.name), func(t *testing.T) {
			got, err := tt.resolver.ToJSON(graphql.ResolveParams{Context: context.Background(), Source: tt.in})
			if (err != nil) != tt.wantErr {
				t.Errorf("corev2PipelineImpl.ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("corev2PipelineImpl.ToJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_corev2types_IsTypeOf(t *testing.T) {
	tests := []struct {
		name     string
		resolver interface {
			IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool
		}
		in   interface{}
		want bool
	}{
		{
			name:     "match",
			resolver: &corev2PipelineImpl{},
			in:       corev2.FixturePipeline("name", "default"),
			want:     true,
		},
		{
			name:     "no match",
			resolver: &corev2PipelineImpl{},
			in:       corev2.FixtureEntity("name"),
			want:     false,
		},
		{
			name:     "match",
			resolver: &roleImpl{},
			in:       corev2.FixtureRole("name", "default"),
			want:     true,
		},
		{
			name:     "no match",
			resolver: &roleImpl{},
			in:       corev2.FixtureEntity("name"),
			want:     false,
		},
		{
			name:     "match",
			resolver: &roleBindingImpl{},
			in:       corev2.FixtureRoleBinding("name", "default"),
			want:     true,
		},
		{
			name:     "no match",
			resolver: &roleBindingImpl{},
			in:       corev2.FixtureEntity("name"),
			want:     false,
		},
		{
			name:     "match",
			resolver: &clusterRoleImpl{},
			in:       corev2.FixtureClusterRole("name"),
			want:     true,
		},
		{
			name:     "no match",
			resolver: &clusterRoleImpl{},
			in:       corev2.FixtureEntity("name"),
			want:     false,
		},
		{
			name:     "match",
			resolver: &clusterRoleBindingImpl{},
			in:       corev2.FixtureClusterRoleBinding("name"),
			want:     true,
		},
		{
			name:     "no match",
			resolver: &clusterRoleBindingImpl{},
			in:       corev2.FixtureEntity("name"),
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T/%s", tt.resolver, tt.name), func(t *testing.T) {
			got := tt.resolver.IsTypeOf(tt.in, graphql.IsTypeOfParams{Context: context.Background()})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("corev2PipelineImpl.ToJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
