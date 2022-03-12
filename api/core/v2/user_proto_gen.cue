package v2

// User describes an authenticated user
#User: {
	username?: string @protobuf(1,string)

	// Password is used to define the cleartext password. It was also previously
	// used to store the hashed password
	password?: string @protobuf(2,string)
	groups?: [...string] @protobuf(3,string)
	disabled?: bool @protobuf(4,bool,#"(gogoproto.jsontag)="disabled""#)

	// PasswordHash is the hashed password, which is safe to display
	passwordHash?: string @protobuf(5,string,name=password_hash)
}
