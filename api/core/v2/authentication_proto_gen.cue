package v2

// Tokens contains the structure for exchanging tokens with the API
#Tokens: {
	// Access token is used by client to make request
	access?: string @protobuf(1,string,#"(gogoproto.jsontag)="access_token""#)

	// ExpiresAt unix timestamp describing when the access token is no longer
	// valid
	expiresAt?: int64 @protobuf(2,int64,name=expires_at,#"(gogoproto.jsontag)="expires_at""#)

	// Refresh token is used by client to request a new access token
	refresh?: string @protobuf(3,string,#"(gogoproto.jsontag)="refresh_token""#)
}
