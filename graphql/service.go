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
