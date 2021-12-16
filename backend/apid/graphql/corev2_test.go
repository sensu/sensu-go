package graphql

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

func Test_corev2PipelineImpl_ID(t *testing.T) {
	tests := []struct {
		name    string
		in      *corev2.Pipeline
		want    string
		wantErr bool
	}{
		{
			name:    "default",
			in:      corev2.FixturePipeline("test", "default"),
			want:    "srn:corev2/pipeline:default:test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &corev2PipelineImpl{}
			got, err := tr.ID(graphql.ResolveParams{Context: context.Background(), Source: tt.in})
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

func Test_corev2PipelineImp_ToJSON(t *testing.T) {
	tests := []struct {
		name    string
		in      *corev2.Pipeline
		want    interface{}
		wantErr bool
	}{
		{
			name:    "default",
			in:      corev2.FixturePipeline("name", "default"),
			want:    types.WrapResource(corev2.FixturePipeline("name", "default")),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &corev2PipelineImpl{}
			got, err := tr.ToJSON(graphql.ResolveParams{Context: context.Background(), Source: tt.in})
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

func Test_corev2PipelineImp_IsTypeOf(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want bool
	}{
		{
			name: "match",
			in:   corev2.FixturePipeline("name", "default"),
			want: true,
		},
		{
			name: "no match",
			in:   corev2.FixtureEntity("name"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &corev2PipelineImpl{}
			got := tr.IsTypeOf(tt.in, graphql.IsTypeOfParams{Context: context.Background()})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("corev2PipelineImpl.ToJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
