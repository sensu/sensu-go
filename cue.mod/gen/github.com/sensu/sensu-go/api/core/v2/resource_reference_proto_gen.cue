package v2

// ResourceReference represents a reference to another resource.
//sensu:nogen
#ResourceReference: {
	// Name is the name of the resource to reference.
	name?: string @protobuf(1,string,#"(gogoproto.jsontag)="name""#,#"(gogoproto.moretags)="yaml: \"name\"""#,name=Name)

	// Type is the name of the data type of the resource to reference.
	type?: string @protobuf(2,string,#"(gogoproto.jsontag)="type""#,#"(gogoproto.moretags)="yaml: \"type\"""#,name=Type)

	// APIVersion is the API version of the resource to reference.
	api_version?: string @protobuf(3,string,#"(gogoproto.jsontag)="api_version""#,#"(gogoproto.moretags)="yaml: \"api_version\"""#,name=APIVersion)
}
