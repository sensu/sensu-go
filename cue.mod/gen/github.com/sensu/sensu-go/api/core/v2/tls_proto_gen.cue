package v2

// TLSOptions holds TLS options that are used across the varying Sensu
// components
#TLSOptions: {
	certFile?:             string @protobuf(1,string,name=cert_file)
	keyFile?:              string @protobuf(2,string,name=key_file)
	trustedCaFile?:        string @protobuf(3,string,#"(gogoproto.customname)="TrustedCAFile""#,name=trusted_ca_file)
	insecure_skip_verify?: bool   @protobuf(4,bool,#"(gogoproto.jsontag)="insecure_skip_verify""#)
	clientAuthType?:       bool   @protobuf(5,bool,name=client_auth_type)
}
