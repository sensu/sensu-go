package graphql

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/stretchr/testify/mock"
)

func Test_selfSubjectAccessReviewImpl_Cani(t *testing.T) {
	tests := []struct {
		name    string
		args    schema.QueryExtensionSelfSubjectAccessReviewCaniFieldResolverArgs
		setup   func(*MockGenericClient)
		want    interface{}
		wantErr bool
	}{
		{
			name: "set meta err",
			args: schema.QueryExtensionSelfSubjectAccessReviewCaniFieldResolverArgs{
				Type: &schema.TypeMetaInput{
					Type:       "CheckConfig",
					ApiVersion: "core/v2",
				},
			},
			setup: func(g *MockGenericClient) {
				g.On("SetTypeMeta", mock.Anything).Return(errors.New("test")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no claims",
			args: schema.QueryExtensionSelfSubjectAccessReviewCaniFieldResolverArgs{
				Type: &schema.TypeMetaInput{
					Type:       "CheckConfig",
					ApiVersion: "core/v2",
				},
				Meta: &schema.ObjectMetaInput{
					Namespace: "default",
					Name:      "sensu",
				},
				Verb: "create",
			},
			setup: func(g *MockGenericClient) {
				g.On("SetTypeMeta", mock.Anything).Return(nil).Once()
				g.On("Authorize", mock.Anything, "create", "sensu").Return(authorization.ErrNoClaims).Once()
			},
			want: map[string]interface{}{
				"code":    "ERR_PERMISSION_DENIED",
				"message": authorization.ErrNoClaims.Error(),
			},
			wantErr: false,
		},
		{
			name: "unauthorized",
			args: schema.QueryExtensionSelfSubjectAccessReviewCaniFieldResolverArgs{
				Type: &schema.TypeMetaInput{
					Type:       "CheckConfig",
					ApiVersion: "core/v2",
				},
				Meta: &schema.ObjectMetaInput{
					Namespace: "default",
					Name:      "sensu",
				},
				Verb: "create",
			},
			setup: func(g *MockGenericClient) {
				g.On("SetTypeMeta", mock.Anything).Return(nil).Once()
				g.On("Authorize", mock.Anything, "create", "sensu").Return(authorization.ErrUnauthorized).Once()
			},
			want: map[string]interface{}{
				"code":    "ERR_PERMISSION_DENIED",
				"message": authorization.ErrUnauthorized.Error(),
			},
			wantErr: false,
		},
		{
			name: "action allowed",
			args: schema.QueryExtensionSelfSubjectAccessReviewCaniFieldResolverArgs{
				Type: &schema.TypeMetaInput{
					Type:       "CheckConfig",
					ApiVersion: "core/v2",
				},
				Meta: &schema.ObjectMetaInput{
					Namespace: "default",
					Name:      "sensu",
				},
				Verb: "create",
			},
			setup: func(g *MockGenericClient) {
				g.On("SetTypeMeta", mock.Anything).Return(nil).Once()
				g.On("Authorize", mock.Anything, "create", "sensu").Return(nil).Once()
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(MockGenericClient)
			params := schema.QueryExtensionSelfSubjectAccessReviewCaniFieldResolverParams{}
			params.Context = context.Background()
			params.Args = tt.args

			// configure client
			tt.setup(client)

			// invoke resolver
			resolver := &selfSubjectAccessReviewImpl{client: client}
			got, err := resolver.Cani(params)
			if (err != nil) != tt.wantErr {
				t.Errorf("selfSubjectAccessReviewImpl.Cani() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selfSubjectAccessReviewImpl.Cani() = %v, want %v", got, tt.want)
			}
		})
	}
}
