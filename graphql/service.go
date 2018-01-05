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
func (service *Service) RegisterEnum(t ScalarDesc) {
	cfg := t.Config()
	registrar := registerTypeWrapper(graphql.NewEnum(cfg))
	service.descRegister[cfg.Name] = registrar
}

// RegisterInput registers a GraphQL type with the service.
func (service *Service) RegisterInput(t InputDesc) {
	cfg := t.Config()
	registrar := func(schema *graphql.Schema) graphql.Type {
		cfg.Fields = inputFieldsThunk(schema, cfg.Fields)
		return graphql.NewInputObject(cfg)
	}
	service.descRegister[cfg.Name] = registrar
}

// RegisterInterface registers a GraphQL type with the service.
func (service *Service) RegisterInterface(t InterfaceDesc, impl InterfaceTypeResolver) {
	cfg := t.Config()
	registrar := func(schema *graphql.Schema) graphql.Type {
		cfg.Fields = fieldsThunk(schema, cfg.Fields)
		cfg.ResolveType = func(p graphql.ResolveTypeParams) *graphql.Object {
			t := impl.ResolveType(p.Value, p)
			return schema.Type(t.Name())
		}
		return graphql.NewInterface(cfg)
	}
	service.descRegister[cfg.Name] = registrar
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

type registerTypeFn func(schema *graphql.Schema) graphql.Type

func registerTypeWrapper(t graphql.Type) registerTypeFn {
	return func(_ *graphql.Schema) graphql.Type {
		return t
	}
}

// Replace mocked types w/ instantiated counterparts
func fieldsThunk(schema *graphql.Schema, fields graphql.Fields) interface{} {
	mockedFields := make([]string, len(fields))
	for _, f := range fields {
		t := unwrapFieldType(f.Type)
		if tt, ok := t.(InputType); ok {
			mockedFields = append(mockedFields, tt.Name())
		}
	}

	if len(fields) == 0 {
		return fields
	}

	return graphql.FieldsThunk(
		func() graphql.Field {
			for _, name := range mockedFields {
				fields[name].Type = schema.Type(name)
			}
			return fields
		},
	)
}

// Replace mocked types w/ instantiated counterparts
func inputFieldsThunk(
	schema *graphql.Schema,
	fields graphql.InputObjectConfigFieldMap,
) interface{} {
	mockedFields := make([]string, len(fields))
	for _, f := range fields {
		t := unwrapFieldType(f.Type)
		if tt, ok := t.(InputType); ok {
			mockedFields = append(mockedFields, tt.Name())
		}
	}

	if len(fields) == 0 {
		return fields
	}

	return graphql.InputObjectConfigFieldMapThunk(
		func() graphql.InputObjectConfigFieldMap {
			for _, name := range mockedFields {
				fields[name].Type = schema.Type(name)
			}
			return fields
		},
	)
}

func unwrapFieldType(t graphql.Type) graphql.Type {
	t := f.Type
	if tt, ok := f.Type.(graphql.NonNull); ok {
		t = tt.OfType
	} else if tt, ok := f.Type.(graphql.List); ok {
		t = tt.OfType
	}
	return t
}
