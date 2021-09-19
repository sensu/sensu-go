package filter

import (
	"context"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/api"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/stretchr/testify/mock"
)

func TestLegacyAdapter_Name(t *testing.T) {
	o := &LegacyAdapter{}
	want := "LegacyAdapter"

	if got := o.Name(); want != got {
		t.Errorf("LegacyAdapter.Name() = %v, want %v", got, want)
	}
}

func TestLegacyAdapter_CanFilter(t *testing.T) {
	type fields struct {
		AssetGetter  asset.Getter
		Store        store.Store
		StoreTimeout time.Duration
	}
	type args struct {
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "returns false when resource reference is not a core/v2.EventFilter",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
				},
			},
			want: false,
		},
		{
			name: "returns false when resource reference is a core/v2.EventFilter and its name is is_incident",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "EventFilter",
					Name:       "is_incident",
				},
			},
			want: false,
		},
		{
			name: "returns false when resource reference is a core/v2.EventFilter and its name is has_metrics",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "EventFilter",
					Name:       "has_metrics",
				},
			},
			want: false,
		},
		{
			name: "returns false when resource reference is a core/v2.EventFilter and its name is not_silenced",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "EventFilter",
					Name:       "not_silenced",
				},
			},
			want: false,
		},
		{
			name: "returns true when resource reference is a core/v2.EventFilter and its name doesn't match a built-in filter",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "EventFilter",
					Name:       "my_unique_filter",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LegacyAdapter{
				AssetGetter:  tt.fields.AssetGetter,
				Store:        tt.fields.Store,
				StoreTimeout: tt.fields.StoreTimeout,
			}
			if got := l.CanFilter(tt.args.ref); got != tt.want {
				t.Errorf("LegacyAdapter.CanFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLegacyAdapter_Filter(t *testing.T) {
	type fields struct {
		AssetGetter  asset.Getter
		Store        store.Store
		StoreTimeout time.Duration
	}
	type args struct {
		ctx   context.Context
		ref   *corev2.ResourceReference
		event *corev2.Event
	}
	newEvent := func() *corev2.Event {
		event := corev2.FixtureEvent("default", "default")
		event.Check.Output = "matched"
		return event
	}
	newFilter := func(action string, expressions []string) *corev2.EventFilter {
		return &corev2.EventFilter{
			ObjectMeta:  corev2.ObjectMeta{Name: "my_filter"},
			Action:      action,
			Expressions: expressions,
		}
	}
	newArgs := func() args {
		return args{
			ctx:   context.Background(),
			ref:   &corev2.ResourceReference{Name: "my_filter"},
			event: newEvent(),
		}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "allow filters deny events that do not match expression",
			args: newArgs(),
			fields: fields{
				Store: func() store.Store {
					filter := newFilter(corev2.EventFilterActionAllow, []string{`event.check.output == "unmatched"`})
					stor := &mockstore.MockStore{}
					stor.On("GetEventFilterByName", mock.Anything, filter.Name).Return(filter, nil)
					return stor
				}(),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "allow filters allow events that match expression",
			args: newArgs(),
			fields: fields{
				Store: func() store.Store {
					filter := newFilter(corev2.EventFilterActionAllow, []string{`event.check.output == "matched"`})
					stor := &mockstore.MockStore{}
					stor.On("GetEventFilterByName", mock.Anything, filter.Name).Return(filter, nil)
					return stor
				}(),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "allow filters deny events that only match some expressions",
			args: newArgs(),
			fields: fields{
				Store: func() store.Store {
					filter := newFilter(corev2.EventFilterActionAllow, []string{
						`event.check.output == "matched"`,
						`event.check.output == "unmatched"`,
					})
					stor := &mockstore.MockStore{}
					stor.On("GetEventFilterByName", mock.Anything, filter.Name).Return(filter, nil)
					return stor
				}(),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "deny filters allow events that do not match expression",
			args: newArgs(),
			fields: fields{
				Store: func() store.Store {
					filter := newFilter(corev2.EventFilterActionDeny, []string{`event.check.output == "unmatched"`})
					stor := &mockstore.MockStore{}
					stor.On("GetEventFilterByName", mock.Anything, filter.Name).Return(filter, nil)
					return stor
				}(),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "deny filters deny events that match expression",
			args: newArgs(),
			fields: fields{
				Store: func() store.Store {
					filter := newFilter(corev2.EventFilterActionDeny, []string{`event.check.output == "matched"`})
					stor := &mockstore.MockStore{}
					stor.On("GetEventFilterByName", mock.Anything, filter.Name).Return(filter, nil)
					return stor
				}(),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "deny filters allow events that only match some expressions",
			args: newArgs(),
			fields: fields{
				Store: func() store.Store {
					filter := newFilter(corev2.EventFilterActionDeny, []string{
						`event.check.output == "unmatched"`,
						`event.check.output == "matched"`,
					})
					stor := &mockstore.MockStore{}
					stor.On("GetEventFilterByName", mock.Anything, filter.Name).Return(filter, nil)
					return stor
				}(),
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LegacyAdapter{
				AssetGetter:  tt.fields.AssetGetter,
				Store:        tt.fields.Store,
				StoreTimeout: tt.fields.StoreTimeout,
			}
			got, err := l.Filter(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("LegacyAdapter.Filter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LegacyAdapter.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_evaluateEventFilter(t *testing.T) {
	type args struct {
		ctx    context.Context
		event  *corev2.Event
		filter *corev2.EventFilter
		assets asset.RuntimeAssetSet
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns false when an event is within a time window with action allow",
			args: args{
				ctx:   context.Background(),
				event: corev2.FixtureEvent("entity1", "check1"),
				filter: func() *corev2.EventFilter {
					now := time.Now().UTC()
					filter := &corev2.EventFilter{
						ObjectMeta: corev2.ObjectMeta{
							Name: "in_time_window_allow",
						},
						Action: corev2.EventFilterActionAllow,
						When: &corev2.TimeWindowWhen{
							Days: corev2.TimeWindowDays{
								All: []*corev2.TimeWindowTimeRange{{
									Begin: now.Add(-time.Minute * time.Duration(1)).Format("03:04PM"),
									End:   now.Add(time.Minute * time.Duration(1)).Format("03:04PM"),
								}},
							},
						},
					}
					return filter
				}(),
			},
			want: false,
		},
		{
			name: "returns true when an event is not within a time window with action allow",
			args: args{
				ctx:   context.Background(),
				event: corev2.FixtureEvent("entity1", "check1"),
				filter: func() *corev2.EventFilter {
					now := time.Now().UTC()
					filter := &corev2.EventFilter{
						ObjectMeta: corev2.ObjectMeta{
							Name: "outside_time_window_allow",
						},
						Action: corev2.EventFilterActionAllow,
						When: &corev2.TimeWindowWhen{
							Days: corev2.TimeWindowDays{
								All: []*corev2.TimeWindowTimeRange{{
									Begin: now.Add(time.Minute * time.Duration(10)).Format("03:04PM"),
									End:   now.Add(time.Minute * time.Duration(20)).Format("03:04PM"),
								}},
							},
						},
					}
					return filter
				}(),
			},
			want: true,
		},
		{
			name: "returns true when an event is within a time window with action deny",
			args: args{
				ctx: context.Background(),
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					return event
				}(),
				filter: func() *corev2.EventFilter {
					now := time.Now().UTC()
					filter := &corev2.EventFilter{
						ObjectMeta: corev2.ObjectMeta{
							Name: "in_time_window_deny",
						},
						Action: corev2.EventFilterActionDeny,
						When: &corev2.TimeWindowWhen{
							Days: corev2.TimeWindowDays{
								All: []*corev2.TimeWindowTimeRange{{
									Begin: now.Add(-time.Minute * time.Duration(1)).Format("03:04PM"),
									End:   now.Add(time.Minute * time.Duration(1)).Format("03:04PM"),
								}},
							},
						},
					}
					return filter
				}(),
			},
			want: true,
		},
		{
			name: "returns false when an event is not within a time window with action deny",
			args: args{
				ctx:   context.Background(),
				event: corev2.FixtureEvent("entity1", "check1"),
				filter: func() *corev2.EventFilter {
					now := time.Now().UTC()
					filter := &corev2.EventFilter{
						ObjectMeta: corev2.ObjectMeta{
							Name: "outside_time_window_deny",
						},
						Action: corev2.EventFilterActionDeny,
						When: &corev2.TimeWindowWhen{
							Days: corev2.TimeWindowDays{
								All: []*corev2.TimeWindowTimeRange{{
									Begin: now.Add(time.Minute * time.Duration(10)).Format("03:04PM"),
									End:   now.Add(time.Minute * time.Duration(20)).Format("03:04PM"),
								}},
							},
						},
					}
					return filter
				}(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := evaluateEventFilter(tt.args.ctx, tt.args.event, tt.args.filter, tt.args.assets); got != tt.want {
				t.Errorf("evaluateEventFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJavascriptStoreAccess(t *testing.T) {
	st := new(mockstore.MockStore)
	pipelineRoleBinding := &corev2.RoleBinding{
		Subjects: []corev2.Subject{
			{
				Type: corev2.GroupType,
				Name: "system:pipeline",
			},
		},
		RoleRef: corev2.RoleRef{
			Type: "Role",
			Name: "system:pipeline",
		},
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "system:pipeline",
		},
	}
	pipelineRole := &corev2.Role{
		Rules: []corev2.Rule{
			{
				Verbs: []string{
					"get", "list",
				},
				Resources: []string{
					"events",
				},
			},
		},
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "system:pipeline",
		},
	}
	event := corev2.FixtureEvent("entity", "check")

	// store mock supports rbac authorizer
	st.On("ListClusterRoleBindings", mock.Anything, mock.Anything).Return(([]*corev2.ClusterRoleBinding)(nil), nil)
	st.On("ListRoleBindings", mock.Anything, mock.Anything).Return([]*corev2.RoleBinding{pipelineRoleBinding}, nil)
	st.On("GetRole", mock.Anything, "system:pipeline").Return(pipelineRole, nil)
	st.On("GetClusterRole", mock.Anything).Return(nil, nil)

	// store mock supports event store
	st.On("GetEventByEntityCheck", mock.Anything, "entity", "check").Return(event, nil)
	st.On("GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).Return((*corev2.Event)(nil), nil)
	st.On("GetEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{event}, nil)

	auth := &rbac.Authorizer{
		Store: st,
	}
	ctx := store.NamespaceContext(context.Background(), "default")
	client := api.NewEventClient(st, auth, nil)
	synthEvent := dynamic.Synthesize(event)
	funcs := map[string]interface{}{
		"FetchEvent": client.FetchEvent,
		"ListEvents": client.ListEvents,
	}
	env := FilterExecutionEnvironment{
		Event: synthEvent,
		Funcs: funcs,
	}
	tests := []struct {
		Name   string
		Expr   string
		Match  bool
		ExpErr string
	}{
		{
			Name: "lookup event, match on status 0",
			Expr: `(function () {
				var e = sensu.FetchEvent("entity", "check");
				return e.check.status == 0;
			})()`,
			Match: true,
		},
		{
			Name: "lookup event, match on status 1",
			Expr: `(function () {
				var e = sensu.FetchEvent("entity", "check");
				return e.check.status == 1;
			})()`,
			Match: false,
		},
		{
			Name: "error - nil event, undefined lookup",
			Expr: `(function () {
				var e = sensu.FetchEvent("batman", "robin");
				return e.check.status == 0;
			})()`,
			ExpErr: "TypeError: Cannot access member 'check' of undefined",
		},
		{
			Name: "list events",
			Expr: `(function () {
				var events = sensu.ListEvents();
				return events[0].check.status == 0;
			})()`,
			Match: true,
		},
		{
			Name: "no access to delete",
			Expr: `(function () {
				sensu.DeleteEvent("entity", "check");
				return true;
			})()`,
			ExpErr: "TypeError: 'DeleteEvent' is not a function",
		},
		{
			Name: "no access to update",
			Expr: `(function () {
				sensu.UpdateEvent({});
				return true;
			})()`,
			ExpErr: "TypeError: 'UpdateEvent' is not a function",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			match, err := env.Eval(ctx, test.Expr)
			if err != nil && test.ExpErr == "" {
				t.Fatal(err)
			}
			if err != nil && test.ExpErr != "" {
				if got, want := err.Error(), test.ExpErr; got != want {
					t.Errorf("bad error: got %q, want %q", got, want)
				}
			}
			if test.ExpErr != "" && err == nil {
				t.Fatal("expected non-nil error")
			}
			if got, want := match, test.Match; got != want {
				t.Errorf("bad eval: got %v, want %v", got, want)
			}
		})
	}
	st.AssertCalled(t, "GetEventByEntityCheck", mock.Anything, "entity", "check")
	st.AssertCalled(t, "GetEventByEntityCheck", mock.Anything, "batman", "robin")
	st.AssertCalled(t, "GetEvents", mock.Anything, mock.Anything)
}
