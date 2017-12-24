package service

// KindCode us an unsigned 16-bit integer used for defining kinds.
type KindCode uint16

const (
	// EnumKind identifies enum config
	EnumKind KindCode = iota
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
	kind int
}

// NewType returns new instance of Type
func NewType(name string, kind KindCode) Type {
	return Type{
		name: name,
		kind: kind,
	}
}

// Name of GraphQL type represented
func (t Type) Name() string {
	t.name
}

// Kind of GraphQL type represented
func (t Type) Kind() KindCode {
	t.kind
}
