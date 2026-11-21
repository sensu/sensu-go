package pipeline

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/proto" //nolint:staticcheck // ignore SA1019
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/pipeline/filter"
	"github.com/sensu/sensu-go/backend/pipeline/handler"
	"github.com/sensu/sensu-go/backend/pipeline/mutator"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockexecutor"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

func init() {
	logrus.SetOutput(io.Discard)
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
		Store           store.Store
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
		Store           store.Store
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
				Store: func() store.Store {
					var pipeline *corev2.Pipeline
					err := &store.ErrInternal{Message: "etcd timeout"}
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, "pipeline1").Return(pipeline, err)
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
				Store: func() store.Store {
					var pipeline *corev2.Pipeline
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, "pipeline1").Return(pipeline, nil)
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "pipeline does not exist",
		},
		{
			name: "returns error when pipeline has no workflows",
			args: args{
				ctx:      context.Background(),
				ref:      corev2.FixturePipelineReference("pipeline1"),
				resource: corev2.FixtureEvent("entity1", "check1"),
			},
			fields: fields{
				Store: func() store.Store {
					pipeline := &corev2.Pipeline{
						ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
						Workflows:  []*corev2.PipelineWorkflow{},
					}
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, pipeline.GetName()).Return(pipeline, nil)
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
					var nilFilter *corev2.EventFilter
					err := &store.ErrInternal{Message: "etcd timeout"}
					stor := &mockstore.MockStore{}
					stor.On("GetEventFilterByName", mock.Anything, mock.Anything).
						Return(nilFilter, err)
					return []FilterAdapter{
						&filter.LegacyAdapter{
							Store: stor,
						},
					}
				}(),
				Store: func() store.Store {
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
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, pipeline.GetName()).Return(pipeline, nil)
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
				Store: func() store.Store {
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
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, pipeline.GetName()).Return(pipeline, nil)
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
					var nilMutator *corev2.Mutator
					err := &store.ErrInternal{Message: "etcd timeout"}
					stor := &mockstore.MockStore{}
					stor.On("GetMutatorByName", mock.Anything, mock.Anything).
						Return(nilMutator, err)
					return []MutatorAdapter{
						&mutator.LegacyAdapter{
							Store: stor,
						},
					}
				}(),
				Store: func() store.Store {
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
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, pipeline.GetName()).Return(pipeline, nil)
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
					var nilHandler *corev2.Handler
					err := &store.ErrInternal{Message: "etcd timeout"}
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, mock.Anything).
						Return(nilHandler, err)
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
				Store: func() store.Store {
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
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, pipeline.GetName()).Return(pipeline, nil)
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
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, storedHandler.GetName()).Return(storedHandler, nil)
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
				Store: func() store.Store {
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
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, pipeline.GetName()).Return(pipeline, nil)
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
	stor := func() store.Store {
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
		stor := &mockstore.MockStore{}
		stor.On("GetPipelineByName", mock.Anything, mock.Anything).Return(pipeline, nil)
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
		Store           store.Store
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
					event.Check.Handlers = []string{"myhandler"}
					return event
				}(),
			},
			fields: fields{
				Store: func() store.Store {
					handler := &corev2.Handler{
						ObjectMeta: corev2.NewObjectMeta("myhandler", "default"),
					}
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, handler.GetName()).
						Return(handler, nil)
					return stor
				}(),
			},
			want: func() *corev2.Pipeline {
				pipeline := &corev2.Pipeline{
					ObjectMeta: corev2.NewObjectMeta("legacy-pipeline", "default"),
					Workflows: []*corev2.PipelineWorkflow{
						{
							Name: "legacy-pipeline-workflow-myhandler",
							Filters: []*corev2.ResourceReference{{
								APIVersion: "core/v2",
								Type:       "EventFilter",
								Name:       "not_silenced",
							}},
							Handler: &corev2.ResourceReference{
								Type:       "Handler",
								APIVersion: "core/v2",
								Name:       "myhandler",
							},
						},
					},
				}
				return pipeline
			}(),
		},
		{
			name: "returns error when ref name is legacy-pipeline but no handlers exist",
			args: args{
				ctx: context.WithValue(context.Background(), corev2.NamespaceKey, "default"),
				ref: &corev2.ResourceReference{
					Name: "legacy-pipeline",
				},
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("entity1", "check1")
					event.Check.Handlers = []string{"myhandler"}
					return event
				}(),
			},
			fields: fields{
				Store: func() store.Store {
					var missingHandler *corev2.Handler
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, mock.Anything).
						Return(missingHandler, nil)
					return stor
				}(),
			},
			wantErr: true,
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
				Store: func() store.Store {
					pipeline := &corev2.Pipeline{
						ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
						Workflows:  []*corev2.PipelineWorkflow{},
					}
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, pipeline.GetName()).
						Return(pipeline, nil)
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
				Store: func() store.Store {
					var pipeline *corev2.Pipeline
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, "pipeline1").
						Return(pipeline, errors.New("store error"))
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "store error",
		},
		{
			name: "returns an error if the pipeline is nil",
			args: args{
				ctx: context.WithValue(context.Background(), corev2.NamespaceKey, "default"),
				ref: &corev2.ResourceReference{
					Name: "pipeline1",
				},
			},
			fields: fields{
				Store: func() store.Store {
					var pipeline *corev2.Pipeline
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, "pipeline1").
						Return(pipeline, nil)
					return stor
				}(),
			},
			wantErr:    true,
			wantErrMsg: "pipeline does not exist",
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
				Store: func() store.Store {
					pipeline := &corev2.Pipeline{
						ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
					}
					stor := &mockstore.MockStore{}
					stor.On("GetPipelineByName", mock.Anything, "pipeline1").
						Return(pipeline, nil)
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
		Store           store.Store
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
				Store: func() store.Store {
					handler := corev2.FixtureHandler("handler1")
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, handler.GetName()).
						Return(handler, nil)
					return stor
				}(),
			},
			want: &corev2.Pipeline{
				ObjectMeta: corev2.NewObjectMeta("legacy-pipeline", "default"),
				Workflows: []*corev2.PipelineWorkflow{{
					Name: "legacy-pipeline-workflow-handler1",
					Filters: []*corev2.ResourceReference{{
						APIVersion: "core/v2",
						Type:       "EventFilter",
						Name:       "not_silenced",
					}},
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
				Store: func() store.Store {
					handler := corev2.FixtureHandler("handler1")
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, handler.GetName()).
						Return(handler, nil)
					return stor
				}(),
			},
			want: &corev2.Pipeline{
				ObjectMeta: corev2.NewObjectMeta("legacy-pipeline", "default"),
				Workflows: []*corev2.PipelineWorkflow{{
					Name: "legacy-pipeline-workflow-handler1",
					Filters: []*corev2.ResourceReference{{
						APIVersion: "core/v2",
						Type:       "EventFilter",
						Name:       "not_silenced",
					}},
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
				Store: func() store.Store {
					checkHandler := corev2.FixtureHandler("checkhandler")
					metricsHandler := corev2.FixtureHandler("metricshandler")
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, checkHandler.GetName()).
						Return(checkHandler, nil)
					stor.On("GetHandlerByName", mock.Anything, metricsHandler.GetName()).
						Return(metricsHandler, nil)
					return stor
				}(),
			},
			want: &corev2.Pipeline{
				ObjectMeta: corev2.NewObjectMeta("legacy-pipeline", "default"),
				Workflows: []*corev2.PipelineWorkflow{
					{
						Name: "legacy-pipeline-workflow-checkhandler",
						Filters: []*corev2.ResourceReference{{
							APIVersion: "core/v2",
							Type:       "EventFilter",
							Name:       "not_silenced",
						}},
						Handler: &corev2.ResourceReference{
							APIVersion: "core/v2",
							Type:       "Handler",
							Name:       "checkhandler",
						},
					},
					{
						Name: "legacy-pipeline-workflow-metricshandler",
						Filters: []*corev2.ResourceReference{{
							APIVersion: "core/v2",
							Type:       "EventFilter",
							Name:       "not_silenced",
						}},
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
		nilHandler *corev2.Handler

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
		Store           store.Store
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
				Store: func() store.Store {
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, "pipeHandler").
						Return(pipeHandler(), nil)
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
				Store: func() store.Store {
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, "pipeHandler").
						Return(nilHandler, &store.ErrInternal{Message: "etcd timeout"})
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
				Store: func() store.Store {
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, "pipeHandler").
						Return(nilHandler, errors.New("error"))
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
				Store: func() store.Store {
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, "setHandler").
						Return(setHandler(), nil)
					stor.On("GetHandlerByName", mock.Anything, "pipeHandler").
						Return(pipeHandler(), nil)
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
				Store: func() store.Store {
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, "recursiveLoopHandler").
						Return(recursiveLoopHandler(), nil)
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
				Store: func() store.Store {
					stor := &mockstore.MockStore{}
					stor.On("GetHandlerByName", mock.Anything, "recursiveLoopHandler").Return(recursiveLoopHandler(), nil)
					stor.On("GetHandlerByName", mock.Anything, "nestedHandler").Return(nestedHandler(), nil)
					stor.On("GetHandlerByName", mock.Anything, "setHandler").Return(setHandler(), nil)
					stor.On("GetHandlerByName", mock.Anything, "pipeHandler").Return(pipeHandler(), nil)
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
			got, err := a.expandHandlers(tt.args.ctx, tt.args.handlers, tt.args.level)
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

type recordHandlerAdapter struct {
	Handled *bool
}

func (recordHandlerAdapter) Name() string {
	return "record_handler_adapter"
}

func (recordHandlerAdapter) CanHandle(*corev2.ResourceReference) bool {
	return true
}

func (r recordHandlerAdapter) Handle(context.Context, *corev2.ResourceReference, *corev2.Event, []byte) error {
	*r.Handled = true
	return nil
}

func TestAdapterV1_Run_SilencedEvents(t *testing.T) {
	tests := []struct {
		name        string
		filterRef   *corev2.ResourceReference
		wantHandled bool
	}{
		{
			name:        "silenced event reaches workflow without filters",
			filterRef:   nil,
			wantHandled: true,
		},
		{
			name: "silenced event is denied by not_silenced filter",
			filterRef: &corev2.ResourceReference{
				APIVersion: "core/v2",
				Type:       "EventFilter",
				Name:       "not_silenced",
			},
			wantHandled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handled bool
			workflow := &corev2.PipelineWorkflow{
				Name: "workflow1",
				Handler: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
					Name:       "handler1",
				},
			}
			if tt.filterRef != nil {
				workflow.Filters = []*corev2.ResourceReference{tt.filterRef}
			}
			pipeline := &corev2.Pipeline{
				ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
				Workflows:  []*corev2.PipelineWorkflow{workflow},
			}
			stor := &mockstore.MockStore{}
			stor.On("GetPipelineByName", mock.Anything, mock.Anything).Return(pipeline, nil)

			event := corev2.FixtureEvent("entity1", "check1")
			event.Check.IsSilenced = true
			event.Check.Silenced = []string{"silence1"}

			a := &AdapterV1{
				Store:           stor,
				FilterAdapters:  []FilterAdapter{&filter.NotSilencedAdapter{}},
				MutatorAdapters: []MutatorAdapter{&mutator.JSONAdapter{}},
				HandlerAdapters: []HandlerAdapter{recordHandlerAdapter{Handled: &handled}},
			}
			if err := a.Run(context.Background(), new(corev2.ResourceReference), event); err != nil {
				t.Fatal(err)
			}
			if handled != tt.wantHandled {
				t.Errorf("handler ran = %v, want %v", handled, tt.wantHandled)
			}
		})
	}
}

func TestAdapterV1_generateLegacyPipeline_NotSilencedFilter(t *testing.T) {
	tests := []struct {
		name        string
		handlerFilt []string
		wantFilters []string
	}{
		{
			name:        "legacy handler without filters gets not_silenced",
			handlerFilt: nil,
			wantFilters: []string{"not_silenced"},
		},
		{
			name:        "legacy handler keeps its own filters and gets not_silenced",
			handlerFilt: []string{"is_incident"},
			wantFilters: []string{"is_incident", "not_silenced"},
		},
		{
			name:        "legacy handler already listing not_silenced is not duplicated",
			handlerFilt: []string{"not_silenced"},
			wantFilters: []string{"not_silenced"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := corev2.FixtureHandler("handler1")
			handler.Filters = tt.handlerFilt
			stor := &mockstore.MockStore{}
			stor.On("GetHandlerByName", mock.Anything, handler.GetName()).Return(handler, nil)

			a := &AdapterV1{Store: stor}
			event := corev2.FixtureEvent("entity1", "check1")
			event.Check.Handlers = []string{"handler1"}

			pipeline, err := a.generateLegacyPipeline(context.Background(), event)
			if err != nil {
				t.Fatal(err)
			}
			if len(pipeline.Workflows) != 1 {
				t.Fatalf("workflows = %d, want 1", len(pipeline.Workflows))
			}
			var got []string
			for _, f := range pipeline.Workflows[0].Filters {
				got = append(got, f.Name)
			}
			if !reflect.DeepEqual(got, tt.wantFilters) {
				t.Errorf("filters = %v, want %v", got, tt.wantFilters)
			}
		})
	}
}

// ---- silence / legacy filter scenario tests ----

// notSilencedRef is the filter reference that legacy handlers get by default.
var notSilencedRef = &corev2.ResourceReference{
	APIVersion: "core/v2",
	Type:       "EventFilter",
	Name:       "not_silenced",
}

var isIncidentRef = &corev2.ResourceReference{
	APIVersion: "core/v2",
	Type:       "EventFilter",
	Name:       "is_incident",
}

func filterNames(refs []*corev2.ResourceReference) []string {
	var names []string
	for _, r := range refs {
		names = append(names, r.Name)
	}
	return names
}

// legacyStore returns a store that knows the given handlers by name. A nil
// value for a name simulates a handler that does not exist.
func legacyStore(handlers map[string]*corev2.Handler) *mockstore.MockStore {
	stor := &mockstore.MockStore{}
	for name, h := range handlers {
		if h == nil {
			stor.On("GetHandlerByName", mock.Anything, name).Return((*corev2.Handler)(nil), nil)
			continue
		}
		stor.On("GetHandlerByName", mock.Anything, name).Return(h, nil)
	}
	return stor
}

func TestLegacyPipelineNotSilencedFilter_Scenarios(t *testing.T) {
	setHandler := func(name string, members ...string) *corev2.Handler {
		h := corev2.FixtureHandler(name)
		h.Type = "set"
		h.Handlers = members
		return h
	}

	tests := []struct {
		name        string
		handlers    map[string]*corev2.Handler
		checkNames  []string
		metricNames []string
		noCheck     bool
		storeErr    error
		wantWf      map[string][]string // workflow name -> filter names
		wantErr     bool
		wantNoWf    bool
	}{
		{
			name:       "handler without filters gets not_silenced",
			handlers:   map[string]*corev2.Handler{"h1": corev2.FixtureHandler("h1")},
			checkNames: []string{"h1"},
			wantWf:     map[string][]string{"legacy-pipeline-workflow-h1": {"not_silenced"}},
		},
		{
			name: "existing filters are kept and not_silenced is appended",
			handlers: func() map[string]*corev2.Handler {
				h := corev2.FixtureHandler("h1")
				h.Filters = []string{"is_incident", "custom"}
				return map[string]*corev2.Handler{"h1": h}
			}(),
			checkNames: []string{"h1"},
			wantWf:     map[string][]string{"legacy-pipeline-workflow-h1": {"is_incident", "custom", "not_silenced"}},
		},
		{
			name: "not_silenced listed first is not duplicated",
			handlers: func() map[string]*corev2.Handler {
				h := corev2.FixtureHandler("h1")
				h.Filters = []string{"not_silenced", "is_incident"}
				return map[string]*corev2.Handler{"h1": h}
			}(),
			checkNames: []string{"h1"},
			wantWf:     map[string][]string{"legacy-pipeline-workflow-h1": {"not_silenced", "is_incident"}},
		},
		{
			name: "not_silenced listed last is not duplicated",
			handlers: func() map[string]*corev2.Handler {
				h := corev2.FixtureHandler("h1")
				h.Filters = []string{"is_incident", "not_silenced"}
				return map[string]*corev2.Handler{"h1": h}
			}(),
			checkNames: []string{"h1"},
			wantWf:     map[string][]string{"legacy-pipeline-workflow-h1": {"is_incident", "not_silenced"}},
		},
		{
			name: "not_silenced listed twice by the handler is left as the handler set it",
			handlers: func() map[string]*corev2.Handler {
				h := corev2.FixtureHandler("h1")
				h.Filters = []string{"not_silenced", "not_silenced"}
				return map[string]*corev2.Handler{"h1": h}
			}(),
			checkNames: []string{"h1"},
			wantWf:     map[string][]string{"legacy-pipeline-workflow-h1": {"not_silenced", "not_silenced"}},
		},
		{
			name: "check and metrics handlers each get not_silenced",
			handlers: map[string]*corev2.Handler{
				"check1":   corev2.FixtureHandler("check1"),
				"metrics1": corev2.FixtureHandler("metrics1"),
			},
			checkNames:  []string{"check1"},
			metricNames: []string{"metrics1"},
			wantWf: map[string][]string{
				"legacy-pipeline-workflow-check1":   {"not_silenced"},
				"legacy-pipeline-workflow-metrics1": {"not_silenced"},
			},
		},
		{
			name: "metrics-only event gets not_silenced on metrics handler",
			handlers: map[string]*corev2.Handler{
				"metrics1": corev2.FixtureHandler("metrics1"),
			},
			metricNames: []string{"metrics1"},
			noCheck:     true,
			wantWf: map[string][]string{
				"legacy-pipeline-workflow-metrics1": {"not_silenced"},
			},
		},
		{
			name: "handler set members each get not_silenced",
			handlers: map[string]*corev2.Handler{
				"set1": setHandler("set1", "m1", "m2"),
				"m1":   corev2.FixtureHandler("m1"),
				"m2":   corev2.FixtureHandler("m2"),
			},
			checkNames: []string{"set1"},
			wantWf: map[string][]string{
				"legacy-pipeline-workflow-m1": {"not_silenced"},
				"legacy-pipeline-workflow-m2": {"not_silenced"},
			},
		},
		{
			name: "missing handler is ignored and others still get not_silenced",
			handlers: map[string]*corev2.Handler{
				"missing": nil,
				"h1":      corev2.FixtureHandler("h1"),
			},
			checkNames: []string{"missing", "h1"},
			wantWf:     map[string][]string{"legacy-pipeline-workflow-h1": {"not_silenced"}},
		},
		{
			name:       "only missing handlers yields no workflows",
			handlers:   map[string]*corev2.Handler{"missing": nil},
			checkNames: []string{"missing"},
			wantNoWf:   true,
		},
		{
			name:       "event without check or metrics yields no workflows",
			handlers:   map[string]*corev2.Handler{},
			noCheck:    true,
			checkNames: nil,
			wantNoWf:   true,
		},
		{
			name:       "transient store error on handler lookup is ignored",
			handlers:   map[string]*corev2.Handler{},
			checkNames: []string{"h1"},
			storeErr:   errors.New("transient"),
			wantNoWf:   true,
		},
		{
			name:       "internal store error is returned to the caller",
			handlers:   map[string]*corev2.Handler{},
			checkNames: []string{"h1"},
			storeErr:   &store.ErrInternal{Message: "etcd down"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stor := legacyStore(tt.handlers)
			if tt.storeErr != nil {
				stor = &mockstore.MockStore{}
				for _, n := range tt.checkNames {
					stor.On("GetHandlerByName", mock.Anything, n).Return((*corev2.Handler)(nil), tt.storeErr)
				}
			}
			a := &AdapterV1{Store: stor}

			event := corev2.FixtureEvent("entity1", "check1")
			if tt.noCheck {
				event.Check = nil
			} else {
				event.Check.Handlers = tt.checkNames
			}
			if len(tt.metricNames) > 0 {
				event.Metrics = corev2.FixtureMetrics()
				event.Metrics.Handlers = tt.metricNames
			}

			pipeline, err := a.generateLegacyPipeline(context.Background(), event)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if tt.wantNoWf {
				if err == nil {
					t.Fatalf("expected no-legacy-handlers error, got pipeline %v", pipeline)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			got := map[string][]string{}
			for _, wf := range pipeline.Workflows {
				got[wf.Name] = filterNames(wf.Filters)
			}
			if !reflect.DeepEqual(got, tt.wantWf) {
				t.Errorf("workflow filters = %v, want %v", got, tt.wantWf)
			}
		})
	}
}

// runSilencedEvent runs an event through a workflow with the given filters and
// reports whether the handler was reached.
func runSilencedEvent(t *testing.T, filters []*corev2.ResourceReference, mutateEvent func(*corev2.Event)) bool {
	t.Helper()
	var handled bool
	workflow := &corev2.PipelineWorkflow{
		Name:    "workflow1",
		Filters: filters,
		Handler: &corev2.ResourceReference{APIVersion: "core/v2", Type: "Handler", Name: "handler1"},
	}
	pipeline := &corev2.Pipeline{
		ObjectMeta: corev2.NewObjectMeta("pipeline1", "default"),
		Workflows:  []*corev2.PipelineWorkflow{workflow},
	}
	stor := &mockstore.MockStore{}
	stor.On("GetPipelineByName", mock.Anything, mock.Anything).Return(pipeline, nil)

	event := corev2.FixtureEvent("entity1", "check1")
	if mutateEvent != nil {
		mutateEvent(event)
	}

	a := &AdapterV1{
		Store:           stor,
		FilterAdapters:  []FilterAdapter{&filter.NotSilencedAdapter{}, &filter.IsIncidentAdapter{}},
		MutatorAdapters: []MutatorAdapter{&mutator.JSONAdapter{}},
		HandlerAdapters: []HandlerAdapter{recordHandlerAdapter{Handled: &handled}},
	}
	if err := a.Run(context.Background(), new(corev2.ResourceReference), event); err != nil {
		t.Fatal(err)
	}
	return handled
}

func silenceEvent(entries ...string) func(*corev2.Event) {
	return func(e *corev2.Event) {
		e.Check.IsSilenced = len(entries) > 0
		e.Check.Silenced = entries
	}
}

func markIncident(e *corev2.Event) {
	e.Check.Status = 2
	e.Check.History = []corev2.CheckHistory{{Status: 2}, {Status: 2}}
}

func TestRun_SilenceMatrix(t *testing.T) {
	tests := []struct {
		name        string
		filters     []*corev2.ResourceReference
		mutateEvent func(*corev2.Event)
		wantHandled bool
	}{
		{
			name:        "not silenced, no filters: handled",
			wantHandled: true,
		},
		{
			name:        "silenced, no filters: handled (silence only applies via not_silenced)",
			mutateEvent: silenceEvent("entity:entity1"),
			wantHandled: true,
		},
		{
			name:        "silenced, not_silenced: suppressed",
			filters:     []*corev2.ResourceReference{notSilencedRef},
			mutateEvent: silenceEvent("*:check1"),
			wantHandled: false,
		},
		{
			name:        "not silenced, not_silenced: handled",
			filters:     []*corev2.ResourceReference{notSilencedRef},
			wantHandled: true,
		},
		{
			name:        "silenced incident, not_silenced then is_incident: suppressed",
			filters:     []*corev2.ResourceReference{notSilencedRef, isIncidentRef},
			mutateEvent: func(e *corev2.Event) { markIncident(e); silenceEvent("*:check1")(e) },
			wantHandled: false,
		},
		{
			name:        "silenced incident, is_incident only: handled (is_incident ignores silences)",
			filters:     []*corev2.ResourceReference{isIncidentRef},
			mutateEvent: func(e *corev2.Event) { markIncident(e); silenceEvent("*:check1")(e) },
			wantHandled: true,
		},
		{
			name:        "silenced non-incident, is_incident only: denied by is_incident",
			filters:     []*corev2.ResourceReference{isIncidentRef},
			mutateEvent: func(e *corev2.Event) { e.Check.Status = 0; silenceEvent("*:check1")(e) },
			wantHandled: false,
		},
		{
			name:        "stale IsSilenced flag with empty Silenced list: not silenced",
			filters:     []*corev2.ResourceReference{notSilencedRef},
			mutateEvent: func(e *corev2.Event) { e.Check.IsSilenced = true; e.Check.Silenced = nil },
			wantHandled: true,
		},
		{
			name:        "Silenced list set but IsSilenced false: suppressed (IsSilenced derives from list)",
			filters:     []*corev2.ResourceReference{notSilencedRef},
			mutateEvent: func(e *corev2.Event) { e.Check.IsSilenced = false; e.Check.Silenced = []string{"*:check1"} },
			wantHandled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runSilencedEvent(t, tt.filters, tt.mutateEvent); got != tt.wantHandled {
				t.Errorf("handled = %v, want %v", got, tt.wantHandled)
			}
		})
	}
}

// Legacy path end to end: the generated legacy pipeline must suppress a silenced
// event and deliver an unsilenced or resolved one.
func TestRun_LegacyHandlerSilenceEndToEnd(t *testing.T) {
	tests := []struct {
		name        string
		mutateEvent func(*corev2.Event)
		wantHandled bool
	}{
		{name: "unsilenced legacy event is delivered", wantHandled: true},
		{name: "silenced legacy event is suppressed", mutateEvent: silenceEvent("*:check1"), wantHandled: false},
		{name: "silence cleared on resolution is delivered", mutateEvent: func(e *corev2.Event) {
			silenceEvent("*:check1")(e)
			silenceEvent()(e)
		}, wantHandled: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handled bool
			stor := &mockstore.MockStore{}
			stor.On("GetHandlerByName", mock.Anything, "handler1").Return(corev2.FixtureHandler("handler1"), nil)

			a := &AdapterV1{
				Store:           stor,
				FilterAdapters:  []FilterAdapter{&filter.NotSilencedAdapter{}},
				MutatorAdapters: []MutatorAdapter{&mutator.JSONAdapter{}},
				HandlerAdapters: []HandlerAdapter{recordHandlerAdapter{Handled: &handled}},
			}
			event := corev2.FixtureEvent("entity1", "check1")
			event.Check.Handlers = []string{"handler1"}
			if tt.mutateEvent != nil {
				tt.mutateEvent(event)
			}
			if err := a.Run(context.Background(), &corev2.ResourceReference{Name: LegacyPipelineName}, event); err != nil {
				t.Fatal(err)
			}
			if handled != tt.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tt.wantHandled)
			}
		})
	}
}

// A silence that has not begun yet must not match the event.
func TestIsSilencedBy_FutureBeginNotEffective(t *testing.T) {
	e := corev2.FixtureEvent("entity1", "check1")
	entry := &corev2.Silenced{
		ObjectMeta: corev2.NewObjectMeta("*:check1", "default"),
		Check:      "check1",
		Begin:      time.Now().Add(time.Hour).Unix(),
	}
	if e.IsSilencedBy(entry) {
		t.Error("silence starting in the future must not match yet")
	}
}
