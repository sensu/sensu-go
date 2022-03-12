package v2

// ObjectMeta is metadata all persisted objects have.
#ObjectMeta: {
	// Name must be unique within a namespace. Name is primarily intended for
	// creation idempotence and configuration definition.
	name?: string @protobuf(1,string,#"(gogoproto.jsontag)="name,omitempty""#,#"(gogoproto.moretags)="yaml: \"name,omitempty\"""#)

	// Namespace defines a logical grouping of objects within which each object
	// name must be unique.
	namespace?: string @protobuf(2,string,#"(gogoproto.jsontag)="namespace,omitempty""#,#"(gogoproto.moretags)="yaml: \"namespace,omitempty\"""#)

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May also be used in filters and token
	// substitution.
	// TODO: Link to Sensu documentation.
	// More info: http://kubernetes.io/docs/user-guide/labels
	labels?: {
		[string]: string
	} @protobuf(3,map[string]string,#"(gogoproto.jsontag)="labels,omitempty""#,#"(gogoproto.moretags)="yaml: \",labels,omitempty\"""#)

	// Annotations is an unstructured key value map stored with a resource that
	// may be set by external tools to store and retrieve arbitrary metadata. They
	// are not queryable and should be preserved when modifying objects.
	// TODO: Link to Sensu documentation.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	annotations?: {
		[string]: string
	} @protobuf(4,map[string]string,#"(gogoproto.jsontag)="annotations,omitempty""#,#"(gogoproto.moretags)="yaml: \"annotations,omitempty\"""#)

	// CreatedBy indicates which user created the resource.
	createdBy?: string @protobuf(5,string,name=created_by,#"(gogoproto.jsontag)="created_by,omitempty""#,#"(gogoproto.moretags)="yaml: \"created_by,omitempty\"""#)
}

// TypeMeta is information that can be used to resolve a data type
#TypeMeta: {
	// Type is the type name of the data type
	Type?: string @protobuf(1,string,#"(gogoproto.jsontag)="type""#,#"(gogoproto.moretags)="yaml: \"type,omitempty\"""#)

	// APIVersion is the APIVersion of the data type
	APIVersion?: string @protobuf(2,string,#"(gogoproto.jsontag)="api_version""#,#"(gogoproto.moretags)="yaml: \"api_version,omitempty\"""#)
}
