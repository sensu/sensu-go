package v2

import (
	"reflect"
	"testing"
)

func TestPipeline_validate(t *testing.T) {
	type fields struct {
		ObjectMeta ObjectMeta
		Workflows  []*PipelineWorkflow
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		wantMsg string
	}{
		{
			name: "fails when name is empty",
			fields: fields{
				ObjectMeta: ObjectMeta{},
			},
			wantErr: true,
			wantMsg: "name must not be empty",
		},
		{
			name: "fails when namespace is empty",
			fields: fields{
				ObjectMeta: ObjectMeta{
					Name: "my-pipeline",
				},
			},
			wantErr: true,
			wantMsg: "namespace must be set",
		},
		{
			name: "fails when a workflow is invalid",
			fields: fields{
				ObjectMeta: ObjectMeta{
					Name:      "my-pipeline",
					Namespace: "default",
				},
				Workflows: []*PipelineWorkflow{{}},
			},
			wantErr: true,
			wantMsg: "workflow name must not be empty",
		},
		{
			name: "succeeds when metadata is valid",
			fields: fields{
				ObjectMeta: ObjectMeta{
					Name:      "my-pipeline",
					Namespace: "default",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipeline{
				ObjectMeta: tt.fields.ObjectMeta,
				Workflows:  tt.fields.Workflows,
			}
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Pipeline.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantMsg {
				t.Errorf("Pipeline.validate() error = %v, wantMsg %v", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestPipelineWorkflow_validate(t *testing.T) {
	type fields struct {
		Name    string
		Filters []*ResourceReference
		Mutator *ResourceReference
		Handler *ResourceReference
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		wantMsg string
	}{
		{
			name:    "fails when name is empty",
			fields:  fields{},
			wantErr: true,
			wantMsg: "name must not be empty",
		},
		{
			name: "fails when filter has missing fields",
			fields: fields{
				Name:    "foo",
				Filters: []*ResourceReference{{}},
			},
			wantErr: true,
			wantMsg: "filter name must not be empty",
		},
		{
			name: "fails when filter cannot filter events",
			fields: fields{
				Name: "foo",
				Filters: []*ResourceReference{
					{
						Name:       "my-filter",
						APIVersion: "core/v2",
						Type:       "Mutator",
					},
				},
			},
			wantErr: true,
			wantMsg: "filter resource type not capable of filtering events: core/v2.Mutator",
		},
		{
			name: "fails when mutator has missing fields",
			fields: fields{
				Name:    "foo",
				Mutator: &ResourceReference{},
			},
			wantErr: true,
			wantMsg: "mutator name must not be empty",
		},
		{
			name: "fails when mutator cannot mutate events",
			fields: fields{
				Name: "foo",
				Mutator: &ResourceReference{
					Name:       "my-mutator",
					APIVersion: "core/v2",
					Type:       "EventFilter",
				},
			},
			wantErr: true,
			wantMsg: "mutator resource type not capable of mutating events: core/v2.EventFilter",
		},
		{
			name: "fails when handler is nil",
			fields: fields{
				Name: "foo",
			},
			wantErr: true,
			wantMsg: "handler must be set",
		},
		{
			name: "fails when handler has missing fields",
			fields: fields{
				Name:    "foo",
				Handler: &ResourceReference{},
			},
			wantErr: true,
			wantMsg: "handler name must not be empty",
		},
		{
			name: "fails when handler cannot handle events",
			fields: fields{
				Name: "foo",
				Handler: &ResourceReference{
					Name:       "my-handler",
					APIVersion: "core/v2",
					Type:       "Mutator",
				},
			},
			wantErr: true,
			wantMsg: "handler resource type not capable of handling events: core/v2.Mutator",
		},
		{
			name: "succeeds when name & handler are set",
			fields: fields{
				Name: "foo",
				Handler: &ResourceReference{
					Name:       "my-handler",
					APIVersion: "core/v2",
					Type:       "Handler",
				},
			},
			wantErr: false,
		},
		{
			name: "succeeds when name, filters, mutator & handler are set",
			fields: fields{
				Name: "foo",
				Filters: []*ResourceReference{
					{
						Name:       "my-filter",
						APIVersion: "core/v2",
						Type:       "EventFilter",
					},
				},
				Mutator: &ResourceReference{
					Name:       "my-mutator",
					APIVersion: "core/v2",
					Type:       "Mutator",
				},
				Handler: &ResourceReference{
					Name:       "my-handler",
					APIVersion: "core/v2",
					Type:       "Handler",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &PipelineWorkflow{
				Name:    tt.fields.Name,
				Filters: tt.fields.Filters,
				Mutator: tt.fields.Mutator,
				Handler: tt.fields.Handler,
			}
			err := w.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PipelineWorkflow.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantMsg {
				t.Errorf("PipelineWorkflow.validate() error = %v, wantMsg %v", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestResourceReference_validate(t *testing.T) {
	type fields struct {
		Name       string
		Type       string
		APIVersion string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		wantMsg string
	}{
		{
			name:    "fails when name is empty",
			fields:  fields{},
			wantErr: true,
			wantMsg: "name must not be empty",
		},
		{
			name: "fails when type is empty",
			fields: fields{
				Name: "foo",
			},
			wantErr: true,
			wantMsg: "type must be set",
		},
		{
			name: "fails when api version is empty",
			fields: fields{
				Name: "foo",
				Type: "bar",
			},
			wantErr: true,
			wantMsg: "api_version must be set",
		},
		{
			name: "succeeds when all fields are set",
			fields: fields{
				Name:       "foo",
				Type:       "bar",
				APIVersion: "v2",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceReference{
				Name:       tt.fields.Name,
				Type:       tt.fields.Type,
				APIVersion: tt.fields.APIVersion,
			}
			err := r.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ResourceReference.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantMsg {
				t.Errorf("ResourceReference.validate() error = %v, wantMsg %v", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestPipelineFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Fielder
		wantKey string
		want    string
	}{
		{
			name:    "exposes name",
			args:    FixturePipeline("contoso", "default"),
			wantKey: "pipeline.name",
			want:    "contoso",
		},
		{
			name:    "exposes namespace",
			args:    FixturePipeline("contoso", "default"),
			wantKey: "pipeline.namespace",
			want:    "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.Fields()
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("Pipeline.Fields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}
