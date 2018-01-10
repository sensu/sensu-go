package graphql

// Kind is an unsigned 16-bit integer used for defining kinds.
type Kind uint16

const (
	// EnumKind identifies enum config
	EnumKind Kind = iota
	// InputKind identifies input object config
	InputKind
	// InterfaceKind identifies interface config
	InterfaceKind
	// ObjectKind identifies object config
	ObjectKind
	// ScalarKind identifies scalar config
	ScalarKind
	// SchemaKind identifies schema config
	SchemaKind
	// UnionKind identifies union config
	UnionKind
)

// Type represents base description of GraphQL type
type Type struct {
	name string
	kind Kind
}

// NewType returns new instance of Type
func NewType(name string, kind Kind) Type {
	return Type{
		name: name,
		kind: kind,
	}
}

// Name of GraphQL type represented
func (t Type) Name() string {
	return t.name
}

// Kind of GraphQL type represented
func (t Type) Kind() Kind {
	return t.kind
}
