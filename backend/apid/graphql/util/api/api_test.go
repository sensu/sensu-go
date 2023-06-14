package util_api

import (
	"fmt"
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/core/v3/types"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
)

func TestUnwrapListResult(t *testing.T) {
	type args struct {
		res interface{}
		err error
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "wrapped results",
			args: args{
				res: []*types.Wrapper{{Value: corev2.FixtureHandler("")}},
				err: nil,
			},
			want:    []interface{}{corev2.FixtureHandler("")},
			wantErr: false,
		},
		{
			name: "unauthorized",
			args: args{
				res: []*types.Wrapper{{Value: corev2.FixtureHandler("")}},
				err: authorization.ErrUnauthorized,
			},
			want:    []interface{}{},
			wantErr: false,
		},
		{
			name: "no claims",
			args: args{
				res: []*types.Wrapper{{Value: corev2.FixtureHandler("")}},
				err: authorization.ErrNoClaims,
			},
			want:    []interface{}{},
			wantErr: false,
		},
		{
			name: "err",
			args: args{
				res: []*types.Wrapper{{Value: corev2.FixtureHandler("")}},
				err: fmt.Errorf("something bad"),
			},
			want:    []interface{}{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnwrapListResult(tt.args.res, tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnwrapListResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnwrapListResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnwrapGetResult(t *testing.T) {
	type args struct {
		res interface{}
		err error
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "standard",
			args: args{
				res: corev2.FixtureHandler(""),
				err: nil,
			},
			want:    corev2.FixtureHandler(""),
			wantErr: false,
		},
		{
			name: "wrapped results",
			args: args{
				res: &types.Wrapper{Value: corev2.FixtureHandler("")},
				err: nil,
			},
			want:    corev2.FixtureHandler(""),
			wantErr: false,
		},
		{
			name: "unauthorized",
			args: args{
				res: &types.Wrapper{Value: corev2.FixtureHandler("")},
				err: authorization.ErrUnauthorized,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "no claims",
			args: args{
				res: &types.Wrapper{Value: corev2.FixtureHandler("")},
				err: authorization.ErrNoClaims,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "store err",
			args: args{
				res: &types.Wrapper{Value: corev2.FixtureHandler("")},
				err: &store.ErrNotFound{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "err",
			args: args{
				res: &types.Wrapper{Value: corev2.FixtureHandler("")},
				err: fmt.Errorf("something bad"),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnwrapGetResult(tt.args.res, tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnwrapGetResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnwrapGetResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleListResult(t *testing.T) {
	type args struct {
		res interface{}
		err error
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "wrapped results",
			args: args{
				res: []*types.Wrapper{{Value: corev2.FixtureHandler("")}},
				err: nil,
			},
			want: map[string]interface{}{
				"nodes": []interface{}{corev2.FixtureHandler("")},
			},
			wantErr: false,
		},
		{
			name: "unauthorized",
			args: args{
				res: []*types.Wrapper{{Value: corev2.FixtureHandler("")}},
				err: authorization.ErrUnauthorized,
			},
			want: map[string]interface{}{
				"code":    "ERR_PERMISSION_DENIED",
				"message": authorization.ErrUnauthorized.Error(),
			},
			wantErr: false,
		},
		{
			name: "no claims",
			args: args{
				res: []*types.Wrapper{{Value: corev2.FixtureHandler("")}},
				err: authorization.ErrNoClaims,
			},
			want: map[string]interface{}{
				"code":    "ERR_PERMISSION_DENIED",
				"message": authorization.ErrNoClaims.Error(),
			},
			wantErr: false,
		},
		{
			name: "err",
			args: args{
				res: []*types.Wrapper{{Value: corev2.FixtureHandler("")}},
				err: fmt.Errorf("something bad"),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleListResult(tt.args.res, tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleListResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleListResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleGetResult(t *testing.T) {
	type args struct {
		res interface{}
		err error
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "standard",
			args: args{
				res: corev2.FixtureHandler(""),
				err: nil,
			},
			want:    corev2.FixtureHandler(""),
			wantErr: false,
		},
		{
			name: "wrapped results",
			args: args{
				res: &types.Wrapper{Value: corev2.FixtureHandler("")},
				err: nil,
			},
			want:    corev2.FixtureHandler(""),
			wantErr: false,
		},
		{
			name: "unauthorized",
			args: args{
				res: &types.Wrapper{Value: corev2.FixtureHandler("")},
				err: authorization.ErrUnauthorized,
			},
			want: map[string]interface{}{
				"code":    "ERR_PERMISSION_DENIED",
				"message": authorization.ErrUnauthorized.Error(),
			},
			wantErr: false,
		},
		{
			name: "no claims",
			args: args{
				res: &types.Wrapper{Value: corev2.FixtureHandler("")},
				err: authorization.ErrNoClaims,
			},
			want: map[string]interface{}{
				"code":    "ERR_PERMISSION_DENIED",
				"message": authorization.ErrNoClaims.Error(),
			},
			wantErr: false,
		},
		{
			name: "store err",
			args: args{
				res: &types.Wrapper{Value: corev2.FixtureHandler("")},
				err: &store.ErrNotFound{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "err",
			args: args{
				res: &types.Wrapper{Value: corev2.FixtureHandler("")},
				err: fmt.Errorf("something bad"),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleGetResult(tt.args.res, tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleGetResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleGetResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapResource(t *testing.T) {
	v3resource := corev3.FixtureEntityConfig("name")
	v2resource := corev2.FixtureAsset("name")
	v2wrapped := types.WrapResource(v2resource)

	tests := []struct {
		name string
		args interface{}
		want types.Wrapper
	}{
		{
			name: "corev3",
			args: v3resource,
			want: types.Wrapper{
				TypeMeta: corev2.TypeMeta{APIVersion: "core/v3", Type: "EntityConfig"},
				Value:    v3resource,
			},
		},
		{
			name: "corev2",
			args: v2resource,
			want: types.Wrapper{
				TypeMeta: corev2.TypeMeta{APIVersion: "core/v2", Type: "Asset"},
				Value:    v2resource,
			},
		},
		{
			name: "wrapped",
			args: types.WrapResource(v2resource),
			want: types.WrapResource(v2resource),
		},
		{
			name: "pointer",
			args: &v2wrapped,
			want: types.WrapResource(v2resource),
		},
		{
			name: "nil",
			args: (*types.Wrapper)(nil),
			want: types.Wrapper{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapResource(tt.args)
			assert.EqualValues(t, tt.want, got)
		})
	}
}
