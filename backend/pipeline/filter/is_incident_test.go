package filter

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestIsIncident_Name(t *testing.T) {
	tests := []struct {
		name string
		i    *IsIncident
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &IsIncident{}
			if got := i.Name(); got != tt.want {
				t.Errorf("IsIncident.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsIncident_CanFilter(t *testing.T) {
	type args struct {
		ctx context.Context
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name string
		i    *IsIncident
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &IsIncident{}
			if got := i.CanFilter(tt.args.ctx, tt.args.ref); got != tt.want {
				t.Errorf("IsIncident.CanFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsIncident_Filter(t *testing.T) {
	type args struct {
		ctx   context.Context
		ref   *corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name    string
		i       *IsIncident
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &IsIncident{}
			got, err := i.Filter(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsIncident.Filter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsIncident.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
