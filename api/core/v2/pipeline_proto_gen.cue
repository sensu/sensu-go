package v2

// Pipeline represents a named collection of pipeline workflows.
#Pipeline: {
	// Metadata contains the name, namespace, labels and annotations of the
	// pipeline.
	Metadata?: #ObjectMeta @protobuf(1,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")

	// Workflows contains one or more pipeline workflows.
	Workflows?: [...#PipelineWorkflow] @protobuf(2,PipelineWorkflow,#"(gogoproto.jsontag)="workflows""#,#"(gogoproto.moretags)="yaml: \"workflows\"""#)
}
