package pipeline

import (
	"context"
	"reflect"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPipelineFilter(t *testing.T) {
	p := &Pipeline{}
	store := &mockstore.MockStore{}
	p.store = store

	// Mock the store responses
	allowFilterBar := &types.EventFilter{
		ObjectMeta: types.ObjectMeta{
			Name: "allowFilterBar",
		},
		Action:      types.EventFilterActionAllow,
		Expressions: []string{`event.check.output == "bar"`},
	}
	allowFilterFoo := &types.EventFilter{
		ObjectMeta: types.ObjectMeta{
			Name: "allowFilterFoo",
		},
		Action:      types.EventFilterActionAllow,
		Expressions: []string{`event.check.output == "foo"`},
	}
	denyFilterBar := &types.EventFilter{
		ObjectMeta: types.ObjectMeta{
			Name: "denyFilterBar",
		},
		Action:      types.EventFilterActionDeny,
		Expressions: []string{`event.check.output == "bar"`},
	}
	denyFilterFoo := &types.EventFilter{
		ObjectMeta: types.ObjectMeta{
			Name: "denyFilterFoo",
		},
		Action:      types.EventFilterActionDeny,
		Expressions: []string{`event.check.output == "foo"`},
	}
	store.On("GetEventFilterByName", mock.Anything, "allowFilterBar").Return(allowFilterBar, nil)
	store.On("GetEventFilterByName", mock.Anything, "allowFilterFoo").Return(allowFilterFoo, nil)
	store.On("GetEventFilterByName", mock.Anything, "denyFilterBar").Return(denyFilterBar, nil)
	store.On("GetEventFilterByName", mock.Anything, "denyFilterFoo").Return(denyFilterFoo, nil)

	testCases := []struct {
		name           string
		status         uint32
		history        []types.CheckHistory
		metrics        *types.Metrics
		silenced       []string
		filters        []string
		expectedFilter string
	}{
		{
			name:           "Silenced With Metrics",
			status:         1,
			metrics:        &types.Metrics{},
			silenced:       []string{"entity1"},
			filters:        []string{"has_metrics"},
			expectedFilter: "",
		},
		{
			name:           "Silenced Without Metrics",
			status:         1,
			metrics:        nil,
			silenced:       []string{"entity1"},
			filters:        []string{"is_incident", "not_silenced"},
			expectedFilter: "not_silenced",
		},
		{
			name:   "Not Transitioned From Incident To Healthy",
			status: 0,
			history: []types.CheckHistory{
				types.CheckHistory{Status: 0},
				types.CheckHistory{Status: 0},
			},
			metrics:        nil,
			silenced:       []string{},
			filters:        []string{"is_incident"},
			expectedFilter: "is_incident",
		},
		{
			name:   "Transitioned From Incident To Healthy",
			status: 0,
			history: []types.CheckHistory{
				types.CheckHistory{Status: 1},
				types.CheckHistory{Status: 0},
			},
			metrics:        nil,
			silenced:       []string{},
			filters:        []string{"is_incident"},
			expectedFilter: "",
		},
	}

	for _, tc := range testCases {
		handler := &types.Handler{
			Type:    "pipe",
			Command: "cat",
			Filters: tc.filters,
		}

		t.Run(tc.name, func(t *testing.T) {
			event := &types.Event{
				Check: &types.Check{
					Status:   tc.status,
					History:  tc.history,
					Output:   "foo",
					Silenced: tc.silenced,
				},
				Entity: &types.Entity{
					ObjectMeta: types.ObjectMeta{
						Namespace: "default",
					},
				},
				Metrics: tc.metrics,
			}

			f, _ := p.processFilters(ctx, refs, event)
			assert.Equal(t, tc.expectedFilter, f)
		})
	}
}

func TestPipelineWhenFilter(t *testing.T) {
	p := &Pipeline{}
	store := &mockstore.MockStore{}
	p.store = store

	event := &types.Event{
		Check: &types.Check{
			Status: 2,
			Output: "bar",
		},
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Namespace: "default",
			},
		},
	}

	testCases := []struct {
		name           string
		filterName     string
		action         string
		begin          time.Duration
		end            time.Duration
		expectedFilter string
	}{
		{
			name:           "in time window action allow",
			filterName:     "in_time_window_allow",
			action:         types.EventFilterActionAllow,
			begin:          -time.Minute * time.Duration(1),
			end:            time.Minute * time.Duration(1),
			expectedFilter: "",
		},
		{
			name:           "outside time window action allow",
			filterName:     "outside_time_window_allow",
			action:         types.EventFilterActionAllow,
			begin:          time.Minute * time.Duration(10),
			end:            time.Minute * time.Duration(20),
			expectedFilter: "outside_time_window_allow",
		},
		{
			name:           "in time window action deny",
			filterName:     "in_time_window_deny",
			action:         types.EventFilterActionDeny,
			begin:          -time.Minute * time.Duration(1),
			end:            time.Minute * time.Duration(1),
			expectedFilter: "in_time_window_deny",
		},
		{
			name:           "outside time window action deny",
			filterName:     "outside_time_window_deny",
			action:         types.EventFilterActionDeny,
			begin:          time.Minute * time.Duration(10),
			end:            time.Minute * time.Duration(20),
			expectedFilter: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now().UTC()

			filter := &types.EventFilter{
				ObjectMeta: types.ObjectMeta{
					Name: tc.name,
				},
				Action:      tc.action,
				Expressions: []string{`event.check.output == "bar"`},
			}

			filter.When = &types.TimeWindowWhen{
				Days: types.TimeWindowDays{
					All: []*types.TimeWindowTimeRange{{
						Begin: now.Add(tc.begin).Format("03:04PM"),
						End:   now.Add(tc.end).Format("03:04PM"),
					}},
				},
			}

			store.On("GetEventFilterByName", mock.Anything, tc.filterName).Return(filter, nil)

			handler := &types.Handler{
				Type:    "pipe",
				Command: "cat",
				Filters: []string{tc.filterName},
			}

			f, _ := p.FilterEvent(handler, event)
			assert.Equal(t, tc.expectedFilter, f)
		})
	}
}

func TestPipeline_getFilterForResource(t *testing.T) {
	type fields struct {
		store                  store.Store
		assetGetter            asset.Getter
		backendEntity          *corev2.Entity
		executor               command.Executor
		storeTimeout           time.Duration
		secretsProviderManager *secrets.ProviderManager
		licenseGetter          licensing.Getter
		filters                []Filter
		mutators               []Mutator
		handlers               []Handler
		runners                []Runner
	}
	type args struct {
		ctx context.Context
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Filter
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipeline{
				store:                  tt.fields.store,
				assetGetter:            tt.fields.assetGetter,
				backendEntity:          tt.fields.backendEntity,
				executor:               tt.fields.executor,
				storeTimeout:           tt.fields.storeTimeout,
				secretsProviderManager: tt.fields.secretsProviderManager,
				licenseGetter:          tt.fields.licenseGetter,
				filters:                tt.fields.filters,
				mutators:               tt.fields.mutators,
				handlers:               tt.fields.handlers,
				runners:                tt.fields.runners,
			}
			got, err := p.getFilterForResource(tt.args.ctx, tt.args.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pipeline.getFilterForResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pipeline.getFilterForResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipeline_processFilters(t *testing.T) {
	type fields struct {
		store                  store.Store
		assetGetter            asset.Getter
		backendEntity          *corev2.Entity
		executor               command.Executor
		storeTimeout           time.Duration
		secretsProviderManager *secrets.ProviderManager
		licenseGetter          licensing.Getter
		filters                []Filter
		mutators               []Mutator
		handlers               []Handler
		runners                []Runner
	}
	type args struct {
		ctx   context.Context
		refs  []*corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipeline{
				store:                  tt.fields.store,
				assetGetter:            tt.fields.assetGetter,
				backendEntity:          tt.fields.backendEntity,
				executor:               tt.fields.executor,
				storeTimeout:           tt.fields.storeTimeout,
				secretsProviderManager: tt.fields.secretsProviderManager,
				licenseGetter:          tt.fields.licenseGetter,
				filters:                tt.fields.filters,
				mutators:               tt.fields.mutators,
				handlers:               tt.fields.handlers,
				runners:                tt.fields.runners,
			}
			got, err := p.processFilters(tt.args.ctx, tt.args.refs, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pipeline.processFilters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Pipeline.processFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}
