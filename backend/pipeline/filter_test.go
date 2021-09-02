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

// func TestPipelineFilter(t *testing.T) {
// 	p := &Pipeline{}
// 	store := &mockstore.MockStore{}
// 	p.store = store

// 	// Mock the store responses
// 	allowFilterBar := &types.EventFilter{
// 		ObjectMeta: types.ObjectMeta{
// 			Name: "allowFilterBar",
// 		},
// 		Action:      types.EventFilterActionAllow,
// 		Expressions: []string{`event.check.output == "bar"`},
// 	}
// 	allowFilterFoo := &types.EventFilter{
// 		ObjectMeta: types.ObjectMeta{
// 			Name: "allowFilterFoo",
// 		},
// 		Action:      types.EventFilterActionAllow,
// 		Expressions: []string{`event.check.output == "foo"`},
// 	}
// 	denyFilterBar := &types.EventFilter{
// 		ObjectMeta: types.ObjectMeta{
// 			Name: "denyFilterBar",
// 		},
// 		Action:      types.EventFilterActionDeny,
// 		Expressions: []string{`event.check.output == "bar"`},
// 	}
// 	denyFilterFoo := &types.EventFilter{
// 		ObjectMeta: types.ObjectMeta{
// 			Name: "denyFilterFoo",
// 		},
// 		Action:      types.EventFilterActionDeny,
// 		Expressions: []string{`event.check.output == "foo"`},
// 	}
// 	store.On("GetEventFilterByName", mock.Anything, "allowFilterBar").Return(allowFilterBar, nil)
// 	store.On("GetEventFilterByName", mock.Anything, "allowFilterFoo").Return(allowFilterFoo, nil)
// 	store.On("GetEventFilterByName", mock.Anything, "denyFilterBar").Return(denyFilterBar, nil)
// 	store.On("GetEventFilterByName", mock.Anything, "denyFilterFoo").Return(denyFilterFoo, nil)

// 	testCases := []struct {
// 		name       string
// 		status     uint32
// 		history    []types.CheckHistory
// 		metrics    *types.Metrics
// 		silenced   []string
// 		filters    []*corev2.ResourceReference
// 		want       bool
// 		wantErr    error
// 		wantErrMsg string
// 	}{
// 		{
// 			name:           "Silenced With Metrics",
// 			status:         1,
// 			metrics:        &types.Metrics{},
// 			silenced:       []string{"entity1"},
// 			filters:        []string{"has_metrics"},
// 			expectedFilter: "",
// 		},
// 		{
// 			name:     "Silenced Without Metrics",
// 			status:   1,
// 			metrics:  nil,
// 			silenced: []string{"entity1"},
// 			filters: []*corev2.ResourceReference{
// 				&corev2.ResourceReference{
// 					APIVersion: "core/v2",
// 					Type:       "EventFilter",
// 					Name:       "is_incident",
// 				},
// 				&corev2.ResourceReference{
// 					APIVersion: "core/v2",
// 					Type:       "EventFilter",
// 					Name:       "not_silenced",
// 				},
// 			},
// 			expectedFilter: "not_silenced",
// 		},
// 		{
// 			name:   "Not Transitioned From Incident To Healthy",
// 			status: 0,
// 			history: []types.CheckHistory{
// 				types.CheckHistory{Status: 0},
// 				types.CheckHistory{Status: 0},
// 			},
// 			metrics:        nil,
// 			silenced:       []string{},
// 			filters:        []string{"is_incident"},
// 			expectedFilter: "is_incident",
// 		},
// 		{
// 			name:   "Transitioned From Incident To Healthy",
// 			status: 0,
// 			history: []types.CheckHistory{
// 				types.CheckHistory{Status: 1},
// 				types.CheckHistory{Status: 0},
// 			},
// 			metrics:        nil,
// 			silenced:       []string{},
// 			filters:        []string{"is_incident"},
// 			expectedFilter: "",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			event := &types.Event{
// 				Check: &types.Check{
// 					Status:   tc.status,
// 					History:  tc.history,
// 					Output:   "foo",
// 					Silenced: tc.silenced,
// 				},
// 				Entity: &types.Entity{
// 					ObjectMeta: types.ObjectMeta{
// 						Namespace: "default",
// 					},
// 				},
// 				Metrics: tc.metrics,
// 			}

// 			f, _ := p.processFilters(ctx, refs, event)
// 			assert.Equal(t, tc.expectedFilter, f)
// 		})
// 	}
// }

// func TestPipelineWhenFilter(t *testing.T) {
// 	p := &Pipeline{}
// 	store := &mockstore.MockStore{}
// 	p.store = store

// 	event := &types.Event{
// 		Check: &types.Check{
// 			Status: 2,
// 			Output: "bar",
// 		},
// 		Entity: &types.Entity{
// 			ObjectMeta: types.ObjectMeta{
// 				Namespace: "default",
// 			},
// 		},
// 	}

// 	testCases := []struct {
// 		name       string
// 		filterName string
// 		action     string
// 		begin      time.Duration
// 		end        time.Duration
// 		want       bool
// 		wantErr    error
// 		wantErrMsg string
// 	}{
// 		{
// 			name:           "in time window action allow",
// 			filterName:     "in_time_window_allow",
// 			action:         types.EventFilterActionAllow,
// 			begin:          -time.Minute * time.Duration(1),
// 			end:            time.Minute * time.Duration(1),
// 			expectedFilter: "",
// 		},
// 		{
// 			name:           "outside time window action allow",
// 			filterName:     "outside_time_window_allow",
// 			action:         types.EventFilterActionAllow,
// 			begin:          time.Minute * time.Duration(10),
// 			end:            time.Minute * time.Duration(20),
// 			expectedFilter: "outside_time_window_allow",
// 		},
// 		{
// 			name:           "in time window action deny",
// 			filterName:     "in_time_window_deny",
// 			action:         types.EventFilterActionDeny,
// 			begin:          -time.Minute * time.Duration(1),
// 			end:            time.Minute * time.Duration(1),
// 			expectedFilter: "in_time_window_deny",
// 		},
// 		{
// 			name:           "outside time window action deny",
// 			filterName:     "outside_time_window_deny",
// 			action:         types.EventFilterActionDeny,
// 			begin:          time.Minute * time.Duration(10),
// 			end:            time.Minute * time.Duration(20),
// 			expectedFilter: "",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			now := time.Now().UTC()

// 			filter := &types.EventFilter{
// 				ObjectMeta: types.ObjectMeta{
// 					Name: tc.name,
// 				},
// 				Action:      tc.action,
// 				Expressions: []string{`event.check.output == "bar"`},
// 			}

// 			filter.When = &types.TimeWindowWhen{
// 				Days: types.TimeWindowDays{
// 					All: []*types.TimeWindowTimeRange{{
// 						Begin: now.Add(tc.begin).Format("03:04PM"),
// 						End:   now.Add(tc.end).Format("03:04PM"),
// 					}},
// 				},
// 			}

// 			store.On("GetEventFilterByName", mock.Anything, tc.filterName).Return(filter, nil)

// 			handler := &types.Handler{
// 				Type:    "pipe",
// 				Command: "cat",
// 				Filters: []string{tc.filterName},
// 			}

// 			f, _ := p.FilterEvent(handler, event)
// 			assert.Equal(t, tc.expectedFilter, f)
// 		})
// 	}
// }

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
