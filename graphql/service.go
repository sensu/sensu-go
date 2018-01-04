package graphql

import "github.com/graphql-go/graphql"

// Service ...TODO...
type Service struct {
	builders []typeBuilder
}

// NewService returns new instance of Service
func NewService() *Service {
	// TODO: opts? logger?
	return &Service{}
}

// RegisterScalar registers a GraphQL type with the service.
func (service *Service) RegisterScalar(t ScalarDesc, impl ScalarResolver) {
	cfg := t.Config()
	cfg.ParseLiteral = impl.ParseLiteral
	cfg.ParseValue = impl.ParseValue
	cfg.Serialize = impl.Serialize

	thunk := thunkifyType(graphql.NewScalar(cfg))
	builder := newTypeBuilder(thunk)
	service.addBuilder(builder)
}

// RegisterEnum registers a GraphQL type with the service.
func (service *Service) RegisterEnum(t ScalarDesc) {
	cfg := t.Config()
	thunk := thunkifyType(graphql.NewEnum(cfg))
	builder := newTypeBuilder(thunk)
	service.addBuilder(builder)
}

// Regenerate generates new schema & executor given registered types.
func (service *Service) Regenerate() error {
	// TODO
	// no-op if not dirty
	// create instances of each scalar
	// create instances of each enum

	// create nil pointer for each object
	// create nil pointer for each interface
	// create nil pointer for each union
	// create nil pointer for each input

	// create instances of each interface
	// create instances of each union
	// create instances of each input
	// create instances of each object
}

func (service *Service) addBuilder(b typeBuilder) {
	service.builders = append(service.builders, b)
}

type typeThunk func() graphql.Type

func thunkifyType(t graphql.Type) typeThunk {
	return func() graphql.Type { return t }
}

type typeBuilder struct {
	fn   typeThunk
	deps []string
}

func newTypeBuilder(fn typeThunk, deps ...string) typeBuilder {
	return typeBuilder{
		fn:   fn,
		deps: deps,
	}
}
