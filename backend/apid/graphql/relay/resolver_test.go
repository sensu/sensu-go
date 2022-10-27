package relay

import (
	"context"
	"errors"
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
)

func TestNodeResolverFindType(t *testing.T) {
	register := NodeRegister{}
	resolver := Resolver{Register: &register}

	register.RegisterResolver(NodeResolver{
		ObjectType: schema.CheckType,
		Translator: globalid.CheckTranslator,
	})

	check := corev2.FixtureCheckConfig("http-check")
	typeID := resolver.FindType(context.Background(), check)
	assert.NotNil(t, typeID)
}

func TestResolver_Find(t *testing.T) {
	tests := []struct {
		name    string
		resolve func(p NodeResolverParams) (interface{}, error)
		gid     string
		want    interface{}
		wantErr bool
	}{
		{
			name:    "bad id",
			gid:     "sdflkasjdflkasjdfl",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unknown id",
			gid:     "srn:checks:cybertruck",
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing",
			gid:  "srn:assets:default:test",
			resolve: func(p NodeResolverParams) (interface{}, error) {
				return nil, nil
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "resolve err",
			gid:  "srn:assets:default:test",
			resolve: func(p NodeResolverParams) (interface{}, error) {
				return nil, errors.New("test")
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success",
			gid:  "srn:assets:default:test",
			resolve: func(p NodeResolverParams) (interface{}, error) {
				return "test", nil
			},
			want:    "test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			register := NodeRegister{}
			resolver := Resolver{Register: &register}

			register.RegisterResolver(NodeResolver{
				ObjectType: schema.AssetType,
				Translator: globalid.AssetTranslator,
				Resolve:    tt.resolve,
			})

			got, err := resolver.Find(context.Background(), tt.gid, graphql.ResolveInfo{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolver.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resolver.Find() = %v, want %v", got, tt.want)
			}
		})
	}
}
