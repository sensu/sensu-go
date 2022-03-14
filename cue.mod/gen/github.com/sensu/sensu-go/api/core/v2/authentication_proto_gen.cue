package v2

// Tokens contains the structure for exchanging tokens with the API
#Tokens: {
	// Access token is used by client to make request
	access_token?: string @protobuf(1,string,#"(gogoproto.jsontag)="access_token""#,name=access)

	// ExpiresAt unix timestamp describing when the access token is no longer
	// valid
	expires_at?: int64 @protobuf(2,int64,#"(gogoproto.jsontag)="expires_at""#)

	// Refresh token is used by client to request a new access token
	refresh_token?: string @protobuf(3,string,#"(gogoproto.jsontag)="refresh_token""#,name=refresh)
}
