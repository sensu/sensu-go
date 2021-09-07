package pipeline

import (
	"context"
	"reflect"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/pipeline/filter"
	"github.com/sensu/sensu-go/backend/store"
)

func TestAdapterV1_processFilters(t *testing.T) {
	type fields struct {
		Store           store.Store
		StoreTimeout    time.Duration
		FilterAdapters  []FilterAdapter
		MutatorAdapters []MutatorAdapter
		HandlerAdapters []HandlerAdapter
	}
	type args struct {
		ctx   context.Context
		refs  []*corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       bool
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns an error when no filter adapters match",
			args: args{
				ctx: context.Background(),
				refs: []*corev2.ResourceReference{
					{
						APIVersion: "core/v2",
						Type:       "EventFilter",
						Name:       "is_incident",
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "no filter adapters were found that can filter the resource: core/v2.EventFilter = is_incident",
		},
		{
			name: "returns true when an event is denied by a filter",
			args: args{
				ctx: context.Background(),
				refs: []*corev2.ResourceReference{
					{
						APIVersion: "core/v2",
						Type:       "EventFilter",
						Name:       "is_incident",
					},
				},
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					return event
				}(),
			},
			fields: fields{
				FilterAdapters: []FilterAdapter{
					&filter.IsIncidentAdapter{},
				},
			},
			want: true,
		},
		{
			name: "returns true when an event is allowed by one filter but denied by another",
			args: args{
				ctx: context.Background(),
				refs: []*corev2.ResourceReference{
					{
						APIVersion: "core/v2",
						Type:       "EventFilter",
						Name:       "is_incident",
					},
					{
						APIVersion: "core/v2",
						Type:       "EventFilter",
						Name:       "has_metrics",
					},
				},
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					event.Check.Status = 1
					return event
				}(),
			},
			fields: fields{
				FilterAdapters: []FilterAdapter{
					&filter.IsIncidentAdapter{},
					&filter.HasMetricsAdapter{},
				},
			},
			want: true,
		},
		{
			name: "returns true when an event is allowed by all filters",
			args: args{
				ctx: context.Background(),
				refs: []*corev2.ResourceReference{
					{
						APIVersion: "core/v2",
						Type:       "EventFilter",
						Name:       "is_incident",
					},
					{
						APIVersion: "core/v2",
						Type:       "EventFilter",
						Name:       "has_metrics",
					},
				},
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					event.Check.Status = 1
					event.Metrics = corev2.FixtureMetrics()
					return event
				}(),
			},
			fields: fields{
				FilterAdapters: []FilterAdapter{
					&filter.HasMetricsAdapter{},
					&filter.IsIncidentAdapter{},
					&filter.NotSilencedAdapter{},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AdapterV1{
				Store:           tt.fields.Store,
				StoreTimeout:    tt.fields.StoreTimeout,
				FilterAdapters:  tt.fields.FilterAdapters,
				MutatorAdapters: tt.fields.MutatorAdapters,
				HandlerAdapters: tt.fields.HandlerAdapters,
			}
			got, err := a.processFilters(tt.args.ctx, tt.args.refs, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.processFilters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErrMsg != err.Error() {
				t.Errorf("AdapterV1.processFilters() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if got != tt.want {
				t.Errorf("AdapterV1.processFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdapterV1_getFilterAdapterForResource(t *testing.T) {
	type fields struct {
		Store           store.Store
		StoreTimeout    time.Duration
		FilterAdapters  []FilterAdapter
		MutatorAdapters []MutatorAdapter
		HandlerAdapters []HandlerAdapter
	}
	type args struct {
		ctx context.Context
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       FilterAdapter
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns an error when no filter adapters match",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "EventFilter",
					Name:       "is_incident",
				},
			},
			wantErr:    true,
			wantErrMsg: "no filter adapters were found that can filter the resource: core/v2.EventFilter = is_incident",
		},
		{
			name: "succeeds when filter adapter matches",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "EventFilter",
					Name:       "is_incident",
				},
			},
			fields: fields{
				FilterAdapters: []FilterAdapter{
					&filter.HasMetricsAdapter{},
					&filter.IsIncidentAdapter{},
					&filter.NotSilencedAdapter{},
				},
			},
			wantErr: false,
			want:    &filter.IsIncidentAdapter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AdapterV1{
				Store:           tt.fields.Store,
				StoreTimeout:    tt.fields.StoreTimeout,
				FilterAdapters:  tt.fields.FilterAdapters,
				MutatorAdapters: tt.fields.MutatorAdapters,
				HandlerAdapters: tt.fields.HandlerAdapters,
			}
			got, err := a.getFilterAdapterForResource(tt.args.ctx, tt.args.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.getFilterAdapterForResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErrMsg != err.Error() {
				t.Errorf("AdapterV1.getFilterAdapterForResource() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AdapterV1.getFilterAdapterForResource() = %v, want %v", got, tt.want)
			}
		})
	}
}
