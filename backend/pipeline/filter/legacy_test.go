package filter

import (
	"context"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestLegacy_Name(t *testing.T) {
	o := &Legacy{}
	want := "Legacy"

	if got := o.Name(); want != got {
		t.Errorf("Legacy.Name() = %v, want %v", got, want)
	}
}

func TestLegacy_CanFilter(t *testing.T) {
	type fields struct {
		AssetGetter  asset.Getter
		Store        store.Store
		StoreTimeout time.Duration
	}
	type args struct {
		ctx context.Context
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
			l := &Legacy{
				AssetGetter:  tt.fields.AssetGetter,
				Store:        tt.fields.Store,
				StoreTimeout: tt.fields.StoreTimeout,
			}
			if got := l.CanFilter(tt.args.ctx, tt.args.ref); got != tt.want {
				t.Errorf("Legacy.CanFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLegacy_Filter(t *testing.T) {
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
			l := &Legacy{
				AssetGetter:  tt.fields.AssetGetter,
				Store:        tt.fields.Store,
				StoreTimeout: tt.fields.StoreTimeout,
			}
			got, err := l.Filter(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Legacy.Filter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Legacy.Filter() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := evaluateEventFilter(tt.args.ctx, tt.args.event, tt.args.filter, tt.args.assets); got != tt.want {
				t.Errorf("evaluateEventFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
