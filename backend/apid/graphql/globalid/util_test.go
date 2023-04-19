package globalid

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/stretchr/testify/assert"
)

func TestStandardDecoder(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	handler := corev2.FixtureHandler("myHandler")
	encoderFn := standardEncoder("handlers", "Name")
	components := encoderFn(ctx, handler)

	assert.Equal("handlers", components.Resource())
	assert.Equal("default", components.Namespace())
	assert.Equal("myHandler", components.UniqueComponent())
}

func Test_GenericTranslator_ForResourceNamed(t *testing.T) {
	type fields struct {
		kind namedResource
		name string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "implicit name",
			fields: fields{
				kind: &corev3.EntityConfig{},
			},
			want: "core/v3.EntityConfig",
		},
		{
			name: "explicit name",
			fields: fields{
				kind: &corev3.EntityConfig{},
				name: "custom_entity_config",
			},
			want: "custom_entity_config",
		},
		{
			name: "implicit name AND doesn't implement typemeta interface",
			fields: fields{
				kind: &corev2.CheckConfig{},
			},
			want: "core/v2.CheckConfig",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenericTranslator(tt.fields.kind, tt.fields.name)
			if got := g.ForResourceNamed(); got != tt.want {
				t.Errorf("genericTranslator.ForResourceNamed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GenericTranslator_IsResponsibleFor(t *testing.T) {
	type fields struct {
		kind namedResource
		name string
	}
	tests := []struct {
		name   string
		fields fields
		arg    interface{}
		want   bool
	}{
		{
			name: "IS responsible",
			fields: fields{
				kind: &corev3.EntityConfig{},
			},
			arg:  corev3.FixtureEntityConfig("test"),
			want: true,
		},
		{
			name: "is NOT responsible",
			fields: fields{
				kind: &corev3.EntityConfig{},
			},
			arg:  corev3.FixtureEntityState("test"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenericTranslator(tt.fields.kind, tt.fields.name)
			if got := g.IsResponsible(tt.arg); got != tt.want {
				t.Errorf("genericTranslator.IsResponsible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GenericTranslator_EncodeToString(t *testing.T) {
	type fields struct {
		kind namedResource
		name string
	}
	type args struct {
		ctx context.Context
		in  interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "implicit name",
			fields: fields{
				kind: &corev3.EntityConfig{},
			},
			args: args{
				ctx: context.Background(),
				in:  corev3.FixtureEntityState("test"),
			},
			want: "srn:core/v3.EntityConfig:default:test",
		},
		{
			name: "implicit name",
			fields: fields{
				kind: &corev3.EntityConfig{},
				name: "custom-name",
			},
			args: args{
				ctx: context.Background(),
				in:  corev3.FixtureEntityState("test"),
			},
			want: "srn:custom-name:default:test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenericTranslator(tt.fields.kind, tt.fields.name)
			if got := g.EncodeToString(tt.args.ctx, tt.args.in); got != tt.want {
				t.Errorf("genericTranslator.EncodeToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
