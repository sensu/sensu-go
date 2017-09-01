package types

// MultitenantResource is a object that belongs to an organization
type MultitenantResource interface {
	Org() string
	Env() string
}
