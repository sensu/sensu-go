package v2

// Asset defines an asset and optionally a list of assets (builds) that agents
// install as a dependency for a check, handler, mutator, etc.
#Asset: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// URL is the location of the asset
	url?: string @protobuf(2,string,#"(gogoproto.customname)="URL""#)

	// Sha512 is the SHA-512 checksum of the asset
	sha512?: string @protobuf(3,string)

	// Filters are a collection of sensu queries, used by the system to
	// determine if the asset should be installed. If more than one filter is
	// present the queries are joined by the "AND" operator.
	filters?: [...string] @protobuf(5,string,#"(gogoproto.jsontag)="filters""#)

	// AssetBuilds are a list of one or more assets to install.
	builds?: [...#AssetBuild] @protobuf(6,AssetBuild,#"(gogoproto.jsontag)="builds""#)

	// Metadata contains the name, namespace, labels and annotations of the
	// asset
	metadata?: #ObjectMeta @protobuf(8,ObjectMeta,"(gogoproto.embed)",#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.nullable)=false")

	// Headers is a collection of key/value string pairs used as HTTP headers
	// for asset retrieval.
	headers?: {
		[string]: string
	} @protobuf(9,map[string]string,#"(gogoproto.jsontag)="headers""#)
}

// AssetBuild defines an individual asset that an asset can install as a
// dependency for a check, handler, mutator, etc.
#AssetBuild: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// URL is the location of the asset
	url?: string @protobuf(2,string,#"(gogoproto.customname)="URL""#)

	// Sha512 is the SHA-512 checksum of the asset
	sha512?: string @protobuf(3,string)

	// Filters are a collection of sensu queries, used by the system to
	// determine if the asset should be installed. If more than one filter is
	// present the queries are joined by the "AND" operator.
	filters?: [...string] @protobuf(5,string,#"(gogoproto.jsontag)="filters""#)

	// Headers is a collection of key/value string pairs used as HTTP headers
	// for asset retrieval.
	headers?: {
		[string]: string
	} @protobuf(9,map[string]string,#"(gogoproto.jsontag)="headers""#)
}
