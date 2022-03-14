package v2

// TessenConfig is the representation of a tessen configuration.
#TessenConfig: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// OptOut is the opt-out status of the tessen configuration
	opt_out?: bool @protobuf(1,bool,#"(gogoproto.jsontag)="opt_out""#)
}
