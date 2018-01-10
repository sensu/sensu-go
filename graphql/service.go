package graphql

import (
	"context"

	"github.com/graphql-go/graphql"
)

// Service ...TODO...
type Service struct {
	types  typeRegister
	schema graphql.Schema
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
	service.types.addType(cfg.Name, ScalarKind, registrar)
}

// RegisterEnum registers a GraphQL type with the service.
func (service *Service) RegisterEnum(t EnumDesc) {
	cfg := t.Config()
	registrar := registerTypeWrapper(graphql.NewEnum(cfg))
	service.types.addType(cfg.Name, EnumKind, registrar)
}

// RegisterInput registers a GraphQL type with the service.
func (service *Service) RegisterInput(t InputDesc) {
	cfg := t.Config()
	registrar := func(m graphql.TypeMap) graphql.Type {
		fields := cfg.Fields.(graphql.InputObjectConfigFieldMap)
		cfg.Fields = inputFieldsThunk(m, fields)
		return graphql.NewInputObject(cfg)
	}
	service.types.addType(cfg.Name, InputKind, registrar)
}

// RegisterInterface registers a GraphQL type with the service.
func (service *Service) RegisterInterface(t InterfaceDesc, impl InterfaceTypeResolver) {
	cfg := t.Config()
	registrar := func(m graphql.TypeMap) graphql.Type {
		cfg.Fields = fieldsThunk(m, cfg.Fields.(graphql.Fields))
		cfg.ResolveType = newResolveTypeFn(m, impl)
		return graphql.NewInterface(cfg)
	}
	service.types.addType(cfg.Name, InterfaceKind, registrar)
}

// RegisterObject registers a GraphQL type with the service.
func (service *Service) RegisterObject(t ObjectDesc, impl interface{}) {
	cfg := t.Config()
	registrar := func(m graphql.TypeMap) graphql.Type {
		fields := cfg.Fields.(graphql.Fields)
		for fieldName, handler := range t.FieldHandlers {
			fields[fieldName].Resolve = handler(impl)
		}

		cfg.Fields = fieldsThunk(m, fields)
		cfg.Interfaces = interfacesThunk(m, cfg.Interfaces)
		cfg.IsTypeOf = newIsTypeOfFn(impl)
		return graphql.NewObject(cfg)
	}
	service.types.addType(cfg.Name, ObjectKind, registrar)
}

// RegisterUnion registers a GraphQL type with the service.
func (service *Service) RegisterUnion(t UnionDesc, impl UnionTypeResolver) {
	cfg := t.Config()
	registrar := func(m graphql.TypeMap) graphql.Type {
		newTypes := make([]*graphql.Object, len(cfg.Types))
		for _, t := range cfg.Types {
			objType := m[t.PrivateName].(*graphql.Object)
			newTypes = append(newTypes, objType)
		}

		cfg.Types = newTypes
		cfg.ResolveType = newResolveTypeFn(m, impl)
		return graphql.NewUnion(cfg)
	}
	service.types.addType(cfg.Name, UnionKind, registrar)
}

// RegisterSchema registers given GraphQL schema with the service.
func (service *Service) RegisterSchema(t SchemaDesc) {
	service.types.setSchema(t)
}

// Regenerate generates new schema given registered types.
func (service *Service) Regenerate() error {
	schema, err := newSchema(service.types)
	if err != nil {
		service.schema = schema
	}
	return err
}

// Do executes request given query string
func (service *Service) Do(
	ctx context.Context,
	q string,
	vars map[string]interface{},
) *graphql.Result {
	params := graphql.Params{
		Schema:         service.schema,
		VariableValues: vars,
		Context:        ctx,
	}
	return graphql.Do(params)
}

type typeRegister struct {
	types  map[Kind]map[string]registerTypeFn
	schema SchemaDesc
}

func (r typeRegister) addType(name string, kind Kind, fn registerTypeFn) {
	r.types[kind][name] = fn
}

func (r typeRegister) setSchema(desc SchemaDesc) {
	r.schema = desc
}

func newSchema(reg typeRegister) (graphql.Schema, error) {
	typeMap := make(graphql.TypeMap, len(reg.types))

	registerTypes(
		typeMap,

		// Register types w/o dependencies first
		reg.types[ScalarKind],
		reg.types[EnumKind],

		// Rest...
		reg.types[InputKind],
		reg.types[ObjectKind],
		reg.types[InterfaceKind],
		reg.types[UnionKind],
	)

	schemaCfg := reg.schema.Config()
	if schemaCfg.Query != nil {
		queryType := typeMap[schemaCfg.Query.Name()]
		schemaCfg.Query = queryType.(*graphql.Object)
	}
	if schemaCfg.Mutation != nil {
		mutationType := typeMap[schemaCfg.Mutation.Name()]
		schemaCfg.Mutation = mutationType.(*graphql.Object)
	}
	if schemaCfg.Subscription != nil {
		subscriptionType := typeMap[schemaCfg.Subscription.Name()]
		schemaCfg.Subscription = subscriptionType.(*graphql.Object)
	}

	return graphql.NewSchema(schemaCfg)
}

func registerTypes(m graphql.TypeMap, col ...map[string]registerTypeFn) {
	for _, fns := range col {
		for name, fn := range fns {
			m[name] = fn(m)
		}
	}
}

type registerTypeFn func(graphql.TypeMap) graphql.Type

func registerTypeWrapper(t graphql.Type) registerTypeFn {
	return func(_ graphql.TypeMap) graphql.Type {
		return t
	}
}
