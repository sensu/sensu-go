package v2

// ResourceReference represents a reference to another resource.
//sensu:nogen
#ResourceReference: {
	// Name is the name of the resource to reference.
	Name?: string @protobuf(1,string,#"(gogoproto.jsontag)="name""#,#"(gogoproto.moretags)="yaml: \"name\"""#)

	// Type is the name of the data type of the resource to reference.
	Type?: string @protobuf(2,string,#"(gogoproto.jsontag)="type""#,#"(gogoproto.moretags)="yaml: \"type\"""#)

	// APIVersion is the API version of the resource to reference.
	APIVersion?: string @protobuf(3,string,#"(gogoproto.jsontag)="api_version""#,#"(gogoproto.moretags)="yaml: \"api_version\"""#)
}
