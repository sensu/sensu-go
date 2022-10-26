package pipeline

import (
	"context"
	"errors"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/proto" //nolint:staticcheck // ignore SA1019
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/pipeline/filter"
	"github.com/sensu/sensu-go/backend/pipeline/handler"
	"github.com/sensu/sensu-go/backend/pipeline/mutator"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockexecutor"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

func TestAdapterV1_Name(t *testing.T) {
	o := &AdapterV1{}
	want := "AdapterV1"

	if got := o.Name(); want != got {
		t.Errorf("AdapterV1.Name() = %v, want %v", got, want)
	}
}

func TestAdapterV1_CanRun(t *testing.T) {
	type fields struct {
		Store           storev2.Interface
		StoreTimeout    time.Duration
		FilterAdapters  []FilterAdapter
		MutatorAdapters []MutatorAdapter
		HandlerAdapters []HandlerAdapter
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
			name: "returns false when resource reference is not a core/v2.Pipeline",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
				},
			},
			want: false,
		},
		{
			name: "returns true when resource reference is a core/v2.Pipeline",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Pipeline",
				},
			},
			want: true,
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
			if got := a.CanRun(tt.args.ref); got != tt.want {
				t.Errorf("AdapterV1.CanRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdapterV1_Run(t *testing.T) {
	type fields struct {
		Store           storev2.Interface
		StoreTimeout    time.Duration
		FilterAdapters  []FilterAdapter
		MutatorAdapters []MutatorAdapter
		HandlerAdapters []HandlerAdapter
	}
	type args struct {
		ctx      context.Context
		ref      *corev2.ResourceReference
		resource interface{}
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns error when resource is not a core/v2.Event",
			args: args{
				resource: corev2.FixtureHandler("handler1"),
				ref:      corev2.FixturePipelineReference("pipeline1"),
			},
			wantErr:    true,
			wantErrMsg: "resource is not a corev2.Event",
		},
		{
			name: "returns error when the store returns an error",
			args: args{
				ctx:      context.Background(),
				ref:      corev2.FixturePipelineReference("pipeline1"),
				resource: corev2.FixtureEvent("entity1", "check1"),
			},
			fields: fields{
				Store: func() storev2.Interface {
					err := &store.ErrInternal{Message: "etcd timeout"}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(nil, err)
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "internal error: etcd timeout",
		},
		{
			name: "returns error when pipeline does not exist",
			args: args{
				ctx:      context.Background(),
				ref:      corev2.FixturePipelineReference("pipeline1"),
				resource: corev2.FixtureEvent("entity1", "check1"),
			},
			fields: fields{
				Store: func() storev2.Interface {
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(nil, &store.ErrNotFound{})
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "key  not found",
		},
		{
			name: "returns error when pipeline has no workflows",
			args: args{
				ctx:      context.Background(),
				ref:      corev2.FixturePipelineReference("pipeline1"),
				resource: corev2.FixtureEvent("entity1", "check1"),
			},
			fields: fields{
				Store: func() storev2.Interface {
					pipeline := &corev2.Pipeline{
						ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
						Workflows:  []*corev2.PipelineWorkflow{},
					}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Pipeline]{Value: pipeline}, nil)
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "pipeline has no workflows",
		},
		{
			name: "returns error when filter produces an error",
			args: args{
				ctx:      context.Background(),
				ref:      corev2.FixturePipelineReference("pipeline1"),
				resource: corev2.FixtureEvent("entity1", "check1"),
			},
			fields: fields{
				FilterAdapters: func() []FilterAdapter {
					err := &store.ErrInternal{Message: "etcd timeout"}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(nil, err)
					return []FilterAdapter{
						&filter.LegacyAdapter{
							Store: stor,
						},
					}
				}(),
				Store: func() storev2.Interface {
					pipeline := &corev2.Pipeline{
						ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
						Workflows: []*corev2.PipelineWorkflow{
							{
								Name: "send metrics to prometheus",
								Filters: []*corev2.ResourceReference{{
									APIVersion: "core/v2",
									Type:       "EventFilter",
									Name:       "filter1",
								}},
							},
						},
					}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Pipeline]{Value: pipeline}, nil)
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "internal error: etcd timeout",
		},
		{
			name: "returns nil when event is filtered",
			args: args{
				ctx:      context.Background(),
				ref:      corev2.FixturePipelineReference("pipeline1"),
				resource: corev2.FixtureEvent("entity1", "check1"),
			},
			fields: fields{
				FilterAdapters: func() []FilterAdapter {
					return []FilterAdapter{
						&filter.HasMetricsAdapter{},
					}
				}(),
				Store: func() storev2.Interface {
					pipeline := &corev2.Pipeline{
						ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
						Workflows: []*corev2.PipelineWorkflow{
							{
								Name: "send metrics to prometheus",
								Filters: []*corev2.ResourceReference{{
									APIVersion: "core/v2",
									Type:       "EventFilter",
									Name:       "has_metrics",
								}},
							},
						},
					}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Pipeline]{Value: pipeline}, nil)
					return stor
				}(),
			},
			wantErr: false,
		},
		{
			name: "returns error when mutator produces an error",
			args: args{
				ctx:      context.Background(),
				ref:      corev2.FixturePipelineReference("pipeline1"),
				resource: corev2.FixtureEvent("entity1", "check1"),
			},
			fields: fields{
				MutatorAdapters: func() []MutatorAdapter {
					err := &store.ErrInternal{Message: "etcd timeout"}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(nil, err)
					return []MutatorAdapter{
						&mutator.LegacyAdapter{
							Store: stor,
						},
					}
				}(),
				Store: func() storev2.Interface {
					pipeline := &corev2.Pipeline{
						ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
						Workflows: []*corev2.PipelineWorkflow{
							{
								Name: "send metrics to prometheus",
								Mutator: &corev2.ResourceReference{
									APIVersion: "core/v2",
									Type:       "Mutator",
									Name:       "mutator1",
								},
							},
						},
					}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Pipeline]{Value: pipeline}, nil)
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "internal error: etcd timeout",
		},
		{
			name: "returns error when handler produces an error",
			args: args{
				ctx:      context.Background(),
				ref:      corev2.FixturePipelineReference("pipeline1"),
				resource: corev2.FixtureEvent("entity1", "check1"),
			},
			fields: fields{
				MutatorAdapters: []MutatorAdapter{
					&mutator.JSONAdapter{},
				},
				HandlerAdapters: func() []HandlerAdapter {
					err := &store.ErrInternal{Message: "etcd timeout"}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(nil, err)
					ex := &mockexecutor.MockExecutor{}
					execution := command.FixtureExecutionResponse(0, "foo")
					ex.Return(execution, nil)
					return []HandlerAdapter{
						&handler.LegacyAdapter{
							Store:    stor,
							Executor: ex,
						},
					}
				}(),
				Store: func() storev2.Interface {
					pipeline := &corev2.Pipeline{
						ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
						Workflows: []*corev2.PipelineWorkflow{
							{
								Name: "send metrics to prometheus",
								Handler: &corev2.ResourceReference{
									APIVersion: "core/v2",
									Type:       "Handler",
									Name:       "handler1",
								},
							},
						},
					}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Pipeline]{Value: pipeline}, nil)
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "failed to fetch handler from store: internal error: etcd timeout",
		},
		{
			name: "returns nil when pipeline successfully runs",
			args: args{
				ctx:      context.Background(),
				ref:      corev2.FixturePipelineReference("pipeline1"),
				resource: corev2.FixtureEvent("entity1", "check1"),
			},
			fields: fields{
				MutatorAdapters: []MutatorAdapter{
					&mutator.JSONAdapter{},
				},
				HandlerAdapters: func() []HandlerAdapter {
					storedHandler := corev2.FixtureHandler("handler1")
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Handler]{Value: storedHandler}, nil)
					ex := &mockexecutor.MockExecutor{}
					execution := command.FixtureExecutionResponse(0, "foo")
					ex.Return(execution, nil)
					return []HandlerAdapter{
						&handler.LegacyAdapter{
							Store:    stor,
							Executor: ex,
						},
					}
				}(),
				Store: func() storev2.Interface {
					pipeline := &corev2.Pipeline{
						ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
						Workflows: []*corev2.PipelineWorkflow{
							{
								Name:    "send metrics to prometheus",
								Filters: nil,
								Mutator: nil,
								Handler: &corev2.ResourceReference{
									APIVersion: "core/v2",
									Type:       "Handler",
									Name:       "handler1",
								},
							},
						},
					}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Pipeline]{Value: pipeline}, nil)
					return stor
				}(),
			},
			wantErr: false,
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
			err := a.Run(tt.args.ctx, tt.args.ref, tt.args.resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("AdapterV1.Run() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}
		})
	}
}

type cancelContextFilterAdapter struct {
	Cancel context.CancelFunc
	Run    *bool
}

func (cancelContextFilterAdapter) Name() string {
	return "cancel_context_filter_adapter"
}

func (cancelContextFilterAdapter) CanFilter(*corev2.ResourceReference) bool {
	return true
}

func (c cancelContextFilterAdapter) Filter(context.Context, *corev2.ResourceReference, *corev2.Event) (bool, error) {
	c.Cancel()
	*c.Run = true
	// returning false means to _not_ filter the event
	return false, context.Canceled
}

type failIfRunHandlerAdapter struct {
	T *testing.T
}

func (failIfRunHandlerAdapter) Name() string {
	return "fail_if_run_handler_adapter"
}

func (failIfRunHandlerAdapter) CanHandle(*corev2.ResourceReference) bool {
	return true
}

func (f failIfRunHandlerAdapter) Handle(context.Context, *corev2.ResourceReference, *corev2.Event, []byte) error {
	f.T.Fatal("handler was run")
	return nil
}

func TestHandlerDoesNotRunAfterFilterContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var filterRun bool
	filterAdapters := []FilterAdapter{
		cancelContextFilterAdapter{
			Run:    &filterRun,
			Cancel: cancel,
		},
	}
	handlerAdapters := []HandlerAdapter{
		failIfRunHandlerAdapter{T: t},
	}
	stor := func() storev2.Interface {
		pipeline := &corev2.Pipeline{
			ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
			Workflows: []*corev2.PipelineWorkflow{
				{
					Name: "send metrics to prometheus",
					Handler: &corev2.ResourceReference{
						APIVersion: "core/v2",
						Type:       "Handler",
						Name:       "handler1",
					},
					Filters: []*corev2.ResourceReference{
						&corev2.ResourceReference{
							APIVersion: "core/v2",
							Type:       "EventFilter",
							Name:       "filter1",
						},
					},
				},
			},
		}
		stor := &mockstore.V2MockStore{}
		stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Pipeline]{Value: pipeline}, nil)
		return stor
	}()
	a := &AdapterV1{
		Store:          stor,
		FilterAdapters: filterAdapters,
		MutatorAdapters: []MutatorAdapter{
			&mutator.JSONAdapter{},
		},
		HandlerAdapters: handlerAdapters,
	}
	if err := a.Run(ctx, new(corev2.ResourceReference), corev2.FixtureEvent("foo", "bar")); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
	} else {
		t.Fatal("no error from adapter Run()")
	}
	if !filterRun {
		t.Fatal("filter was never run")
	}
}

func TestAdapterV1_resolvePipelineReference(t *testing.T) {
	type fields struct {
		Store           storev2.Interface
		StoreTimeout    time.Duration
		FilterAdapters  []FilterAdapter
		MutatorAdapters []MutatorAdapter
		HandlerAdapters []HandlerAdapter
	}
	type args struct {
		ctx   context.Context
		ref   *corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *corev2.Pipeline
		wantErr bool
	}{
		{
			name: "returns a legacy pipeline if the ref name is legacy-pipeline",
			args: args{
				ctx: context.WithValue(context.Background(), corev2.NamespaceKey, "default"),
				ref: &corev2.ResourceReference{
					Name: "legacy-pipeline",
				},
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					return event
				}(),
			},
			want: func() *corev2.Pipeline {
				pipeline := &corev2.Pipeline{
					ObjectMeta: corev2.NewObjectMeta("legacy-pipeline", "default"),
					Workflows:  []*corev2.PipelineWorkflow{},
				}
				return pipeline
			}(),
		},
		{
			name: "returns a stored pipeline if the ref name is not legacy-pipeline",
			args: args{
				ctx: context.WithValue(context.Background(), corev2.NamespaceKey, "default"),
				ref: &corev2.ResourceReference{
					Name: "pipeline1",
				},
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					return event
				}(),
			},
			fields: fields{
				Store: func() storev2.Interface {
					pipeline := &corev2.Pipeline{
						ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
						Workflows:  []*corev2.PipelineWorkflow{},
					}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Pipeline]{Value: pipeline}, nil)
					return stor
				}(),
			},
			want: func() *corev2.Pipeline {
				pipeline := &corev2.Pipeline{
					ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
					Workflows:  []*corev2.PipelineWorkflow{},
				}
				return pipeline
			}(),
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
			got, err := a.resolvePipelineReference(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.resolvePipelineReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !proto.Equal(got, tt.want) {
				t.Errorf("AdapterV1.resolvePipelineReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdapterV1_getPipelineFromStore(t *testing.T) {
	type fields struct {
		Store           storev2.Interface
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
		want       *corev2.Pipeline
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns an error if the store returns an error",
			args: args{
				ctx: context.WithValue(context.Background(), corev2.NamespaceKey, "default"),
				ref: &corev2.ResourceReference{
					Name: "pipeline1",
				},
			},
			fields: fields{
				Store: func() storev2.Interface {
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(nil, errors.New("store error"))
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "store error",
		},
		{
			name: "returns an error if the pipeline is not found",
			args: args{
				ctx: context.WithValue(context.Background(), corev2.NamespaceKey, "default"),
				ref: &corev2.ResourceReference{
					Name: "pipeline1",
				},
			},
			fields: fields{
				Store: func() storev2.Interface {
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(nil, &store.ErrNotFound{})
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "key  not found",
		},
		{
			name: "returns a pipeline when successful",
			args: args{
				ctx: context.WithValue(context.Background(), corev2.NamespaceKey, "default"),
				ref: &corev2.ResourceReference{
					Name: "pipeline1",
				},
			},
			fields: fields{
				Store: func() storev2.Interface {
					pipeline := &corev2.Pipeline{
						ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
					}
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Pipeline]{Value: pipeline}, nil)
					return stor
				}(),
			},
			want: func() *corev2.Pipeline {
				pipeline := &corev2.Pipeline{
					ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
				}
				return pipeline
			}(),
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
			got, err := a.getPipelineFromStore(tt.args.ctx, tt.args.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.getPipelineFromStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("AdapterV1.getPipelineFromStore() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if !proto.Equal(got, tt.want) {
				t.Errorf("AdapterV1.getPipelineFromStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdapterV1_generateLegacyPipeline(t *testing.T) {
	type fields struct {
		Store           storev2.Interface
		StoreTimeout    time.Duration
		FilterAdapters  []FilterAdapter
		MutatorAdapters []MutatorAdapter
		HandlerAdapters []HandlerAdapter
	}
	type args struct {
		ctx   context.Context
		event *corev2.Event
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *corev2.Pipeline
		wantErr bool
	}{
		{
			name: "a legacy pipeline can be generated from an event with check handlers",
			args: args{
				ctx: context.Background(),
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					event.Check.Handlers = []string{"handler1"}
					return event
				}(),
			},
			fields: fields{
				Store: func() storev2.Interface {
					handler := corev2.FixtureHandler("handler1")
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Handler]{Value: handler}, nil)
					return stor
				}(),
			},
			want: &corev2.Pipeline{
				ObjectMeta: corev2.NewObjectMeta("legacy-pipeline", "default"),
				Workflows: []*corev2.PipelineWorkflow{{
					Name: "legacy-pipeline-workflow-handler1",
					Handler: &corev2.ResourceReference{
						APIVersion: "core/v2",
						Type:       "Handler",
						Name:       "handler1",
					},
				}},
			},
		},
		{
			name: "a legacy pipeline can be generated from an event with metrics handlers",
			args: args{
				ctx: context.Background(),
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					event.Metrics = &corev2.Metrics{
						Handlers: []string{"handler1"},
					}
					return event
				}(),
			},
			fields: fields{
				Store: func() storev2.Interface {
					handler := corev2.FixtureHandler("handler1")
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Handler]{Value: handler}, nil)
					return stor
				}(),
			},
			want: &corev2.Pipeline{
				ObjectMeta: corev2.NewObjectMeta("legacy-pipeline", "default"),
				Workflows: []*corev2.PipelineWorkflow{{
					Name: "legacy-pipeline-workflow-handler1",
					Handler: &corev2.ResourceReference{
						APIVersion: "core/v2",
						Type:       "Handler",
						Name:       "handler1",
					},
				}},
			},
		},
		{
			name: "a legacy pipeline can be generated from an event with both check & metrics handlers",
			args: args{
				ctx: context.Background(),
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					event.Check.Handlers = []string{"checkhandler"}
					event.Metrics = &corev2.Metrics{
						Handlers: []string{"metricshandler"},
					}
					return event
				}(),
			},
			fields: fields{
				Store: func() storev2.Interface {
					checkHandler := corev2.FixtureHandler("checkhandler")
					metricsHandler := corev2.FixtureHandler("metricshandler")
					checkReq := storev2.NewResourceRequestFromResource(checkHandler)
					metricsReq := storev2.NewResourceRequestFromResource(metricsHandler)
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, checkReq).Return(mockstore.Wrapper[*corev2.Handler]{Value: checkHandler}, nil)
					stor.On("Get", mock.Anything, metricsReq).Return(mockstore.Wrapper[*corev2.Handler]{Value: metricsHandler}, nil)
					return stor
				}(),
			},
			want: &corev2.Pipeline{
				ObjectMeta: corev2.NewObjectMeta("legacy-pipeline", "default"),
				Workflows: []*corev2.PipelineWorkflow{
					{
						Name: "legacy-pipeline-workflow-checkhandler",
						Handler: &corev2.ResourceReference{
							APIVersion: "core/v2",
							Type:       "Handler",
							Name:       "checkhandler",
						},
					},
					{
						Name: "legacy-pipeline-workflow-metricshandler",
						Handler: &corev2.ResourceReference{
							APIVersion: "core/v2",
							Type:       "Handler",
							Name:       "metricshandler",
						},
					},
				},
			},
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
			got, err := a.generateLegacyPipeline(tt.args.ctx, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.generateLegacyPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !proto.Equal(got, tt.want) {
				t.Errorf("AdapterV1.generateLegacyPipeline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdapterV1_expandHandlers(t *testing.T) {
	var (
		pipeHandler = func() *corev2.Handler {
			return corev2.FixtureHandler("pipeHandler")
		}

		setHandler = func() *corev2.Handler {
			return &corev2.Handler{
				ObjectMeta: corev2.NewObjectMeta("setHandler", "default"),
				Type:       corev2.HandlerSetType,
				Handlers:   []string{"pipeHandler"},
			}
		}

		nestedHandler = func() *corev2.Handler {
			return &corev2.Handler{
				ObjectMeta: corev2.NewObjectMeta("nestedHandler", "default"),
				Type:       corev2.HandlerSetType,
				Handlers:   []string{"setHandler"},
			}
		}

		recursiveLoopHandler = func() *corev2.Handler {
			return &corev2.Handler{
				ObjectMeta: corev2.NewObjectMeta("recursiveLoopHandler", "default"),
				Type:       corev2.HandlerSetType,
				Handlers:   []string{"recursiveLoopHandler"},
			}
		}
	)
	type fields struct {
		Store           storev2.Interface
		StoreTimeout    time.Duration
		FilterAdapters  []FilterAdapter
		MutatorAdapters []MutatorAdapter
		HandlerAdapters []HandlerAdapter
	}
	type args struct {
		ctx      context.Context
		handlers []string
		level    int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       HandlerMap
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "supports pipe handlers",
			args: args{
				ctx:      context.Background(),
				handlers: []string{"pipeHandler"},
			},
			fields: fields{
				Store: func() storev2.Interface {
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Handler]{Value: pipeHandler()}, nil)
					return stor
				}(),
			},
			want: map[string]*corev2.Handler{
				"pipeHandler": pipeHandler(),
			},
		},
		{
			name: "returns an error when an internal store error occurs",
			args: args{
				ctx:      context.Background(),
				handlers: []string{"pipeHandler"},
			},
			fields: fields{
				Store: func() storev2.Interface {
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(nil, &store.ErrInternal{Message: "etcd timeout"})
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "internal error: etcd timeout",
		},
		{
			name: "returns handlers when a non-internal store error occurs",
			args: args{
				ctx:      context.Background(),
				handlers: []string{"pipeHandler"},
			},
			fields: fields{
				Store: func() storev2.Interface {
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
					return stor
				}(),
			},
			want: map[string]*corev2.Handler{},
		},
		{
			name: "supports & expands set handlers",
			args: args{
				ctx:      context.Background(),
				handlers: []string{"setHandler"},
			},
			fields: fields{
				Store: func() storev2.Interface {
					stor := &mockstore.V2MockStore{}
					setReq := storev2.NewResourceRequestFromResource(setHandler())
					pipeReq := storev2.NewResourceRequestFromResource(pipeHandler())
					stor.On("Get", mock.Anything, setReq).Return(mockstore.Wrapper[*corev2.Handler]{Value: setHandler()}, nil)
					stor.On("Get", mock.Anything, pipeReq).Return(mockstore.Wrapper[*corev2.Handler]{Value: pipeHandler()}, nil)
					return stor
				}(),
			},
			want: map[string]*corev2.Handler{
				"pipeHandler": pipeHandler(),
			},
		},
		{
			name: "skips expanding any sets that are nested too deeply",
			args: args{
				ctx:      context.Background(),
				handlers: []string{"recursiveLoopHandler"},
			},
			fields: fields{
				Store: func() storev2.Interface {
					stor := &mockstore.V2MockStore{}
					stor.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Handler]{Value: recursiveLoopHandler()}, nil)
					return stor
				}(),
			},
			want: map[string]*corev2.Handler{},
		},
		{
			name: "supports multiple nested set handlers",
			args: args{
				ctx:      context.Background(),
				handlers: []string{"recursiveLoopHandler", "nestedHandler"},
			},
			fields: fields{
				Store: func() storev2.Interface {
					stor := &mockstore.V2MockStore{}
					recReq := storev2.NewResourceRequestFromResource(recursiveLoopHandler())
					nestReq := storev2.NewResourceRequestFromResource(nestedHandler())
					setReq := storev2.NewResourceRequestFromResource(setHandler())
					pipeReq := storev2.NewResourceRequestFromResource(pipeHandler())
					stor.On("Get", mock.Anything, recReq).Return(mockstore.Wrapper[*corev2.Handler]{Value: recursiveLoopHandler()}, nil)
					stor.On("Get", mock.Anything, nestReq).Return(mockstore.Wrapper[*corev2.Handler]{Value: nestedHandler()}, nil)
					stor.On("Get", mock.Anything, setReq).Return(mockstore.Wrapper[*corev2.Handler]{Value: setHandler()}, nil)
					stor.On("Get", mock.Anything, pipeReq).Return(mockstore.Wrapper[*corev2.Handler]{Value: pipeHandler()}, nil)
					return stor
				}(),
			},
			want: map[string]*corev2.Handler{
				"pipeHandler": pipeHandler(),
			},
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
			got, err := a.expandHandlers(tt.args.ctx, "default", tt.args.handlers, tt.args.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.expandHandlers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("AdapterV1.expandHandlers() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AdapterV1.expandHandlers() = %v, want %v", got, tt.want)
			}
		})
	}
}
