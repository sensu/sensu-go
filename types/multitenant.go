package types

// MultitenantResource is a object that belongs to an organization
type MultitenantResource interface {
	GetOrg() string
	GetEnv() string
}
