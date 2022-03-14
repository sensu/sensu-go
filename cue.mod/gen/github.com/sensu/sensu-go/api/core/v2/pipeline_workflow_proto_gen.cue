package v2

// PipelineWorkflow represents a workflow of filters, mutator, & handler to use
// in a pipeline.
//sensu:nogen
#PipelineWorkflow: {
	// Name is a descriptive name of the pipeline workflow.
	name?: string @protobuf(1,string,#"(gogoproto.jsontag)="name""#,#"(gogoproto.moretags)="yaml: \"name""#,name=Name)

	// Filters contains one or more references to a resource to use as an event
	// filter.
	filters?: [...#ResourceReference] @protobuf(2,ResourceReference,#"(gogoproto.jsontag)="filters,omitempty""#,#"(gogoproto.moretags)="yaml: \"filters,omitempty\"""#,name=Filters)

	// Mutator contains a reference to a resource to use as an event mutator.
	mutator?: #ResourceReference @protobuf(3,ResourceReference,#"(gogoproto.jsontag)="mutator,omitempty""#,#"(gogoproto.moretags)="yaml: \"mutator,omitempty\"""#,name=Mutator)

	// Handler contains a reference to a resource to use as an event handler.
	handler?: #ResourceReference @protobuf(4,ResourceReference,#"(gogoproto.jsontag)="handler""#,#"(gogoproto.moretags)="yaml: \"handler]\"""#,name=Handler)
}
