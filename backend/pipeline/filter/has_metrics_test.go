package filter

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestHasMetrics_Name(t *testing.T) {
	tests := []struct {
		name string
		i    *HasMetrics
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &HasMetrics{}
			if got := i.Name(); got != tt.want {
				t.Errorf("HasMetrics.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasMetrics_CanFilter(t *testing.T) {
	type args struct {
		ctx context.Context
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name string
		i    *HasMetrics
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &HasMetrics{}
			if got := i.CanFilter(tt.args.ctx, tt.args.ref); got != tt.want {
				t.Errorf("HasMetrics.CanFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasMetrics_Filter(t *testing.T) {
	type args struct {
		ctx   context.Context
		ref   *corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name    string
		i       *HasMetrics
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &HasMetrics{}
			got, err := i.Filter(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasMetrics.Filter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HasMetrics.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
