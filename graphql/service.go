package graphql

import "github.com/graphql-go/graphql"

// Service ...TODO...
type Service struct {
	descRegister map[string]registerTypeFn
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

	registrar := registerTypeWrapper(graphql.NewScalar(cfg))
	service.descRegister[cfg.Name] = registrar
}

// RegisterEnum registers a GraphQL type with the service.
func (service *Service) RegisterEnum(t EnumDesc) {
	cfg := t.Config()
	registrar := registerTypeWrapper(graphql.NewEnum(cfg))
	service.descRegister[cfg.Name] = registrar
}

// RegisterInput registers a GraphQL type with the service.
func (service *Service) RegisterInput(t InputDesc) {
	cfg := t.Config()
	registrar := func(schema *graphql.Schema) graphql.Type {
		fields := cfg.Fields.(graphql.InputObjectConfigFieldMap)
		cfg.Fields = inputFieldsThunk(schema, fields)
		return graphql.NewInputObject(cfg)
	}
	service.descRegister[cfg.Name] = registrar
}

// RegisterInterface registers a GraphQL type with the service.
func (service *Service) RegisterInterface(t InterfaceDesc, impl InterfaceTypeResolver) {
	cfg := t.Config()
	registrar := func(schema *graphql.Schema) graphql.Type {
		cfg.Fields = fieldsThunk(schema, cfg.Fields.(graphql.Fields))
		cfg.ResolveType = func(p graphql.ResolveTypeParams) *graphql.Object {
			typeRef := impl.ResolveType(p.Value, p)
			objType := schema.Type(typeRef.Name())
			return objType.(*graphql.Object)
		}
		return graphql.NewInterface(cfg)
	}
	service.descRegister[cfg.Name] = registrar
}

// RegisterUnion registers a GraphQL type with the service.
func (service *Service) RegisterUnion(t UnionDesc, impl UnionTypeResolver) {
	cfg := t.Config()
	registrar := func(schema *graphql.Schema) graphql.Type {
		cfg.ResolveType = func(p graphql.ResolveTypeParams) *graphql.Object {
			typeRef := impl.ResolveType(p.Value, p)
			objType := schema.Type(typeRef.Name())
			return objType.(*graphql.Object)
		}

		newTypes := make([]*graphql.Object, len(cfg.Types))
		for _, t := range cfg.Types {
			objType := schema.Type(t.PrivateName).(*graphql.Object)
			newTypes = append(newTypes, objType)
		}

		cfg.Types = newTypes
		return graphql.NewUnion(cfg)
	}
	service.descRegister[cfg.Name] = registrar
}

// Regenerate generates new schema & executor given registered types.
func (service *Service) Regenerate() error {
	// no-op if not dirty

	// create schema instance?
	// create instances of each scalar
	// create instances of each enum

	// create nil pointer for each object
	// create nil pointer for each interface
	// create nil pointer for each union
	// create nil pointer for each input

	// create instances of each interface
	// create instances of each input
	// create instances of each object
	// create instances of each union

	return nil
}

type registerTypeFn func(schema *graphql.Schema) graphql.Type

func registerTypeWrapper(t graphql.Type) registerTypeFn {
	return func(_ *graphql.Schema) graphql.Type {
		return t
	}
}
