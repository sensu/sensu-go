package graphql

import (
	"context"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// Service describes the whole of a GraphQL schema, validation, and execution.
type Service struct {
	// Executor evaluates a given request and returns a result. If none
	// the default executor is used.
	Executor func(p graphql.ExecuteParams) *graphql.Result

	schema graphql.Schema
	types  *typeRegister
	mware  []Middleware
}

// NewService returns new instance of Service
func NewService() *Service {
	return &Service{
		Executor: graphql.Execute,
		types:    newTypeRegister(),
	}
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
		cfg = t.Config()
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
		cfg = t.Config()
		cfg.ResolveType = nil
		if impl != nil {
			cfg.ResolveType = newResolveTypeFn(m, impl)
		}
		cfg.Fields = fieldsThunk(m, cfg.Fields.(graphql.Fields))
		return graphql.NewInterface(cfg)
	}
	service.types.addType(cfg.Name, InterfaceKind, registrar)
}

// RegisterObject registers a GraphQL type with the service.
func (service *Service) RegisterObject(t ObjectDesc, impl interface{}) {
	cfg := t.Config()
	registrar := func(m graphql.TypeMap) graphql.Type {
		cfg = t.Config()
		fields := cfg.Fields.(graphql.Fields)
		for fieldName, handler := range t.FieldHandlers {
			fields[fieldName].Resolve = handler(impl)
		}

		cfg.IsTypeOf = nil
		if typeResolver, ok := impl.(isTypeOfResolver); ok {
			cfg.IsTypeOf = newIsTypeOfFn(typeResolver)
		}

		for _, ext := range service.types.extensionsForType(cfg.Name) {
			extObjCfg := ext.(graphql.ObjectConfig)
			mergeObjectConfig(cfg, extObjCfg)
		}

		cfg.Fields = fieldsThunk(m, fields)
		cfg.Interfaces = interfacesThunk(m, cfg.Interfaces)
		return graphql.NewObject(cfg)
	}
	service.types.addType(cfg.Name, ObjectKind, registrar)
}

// RegisterObjectExtension registers a GraphQL type with the service.
func (service *Service) RegisterObjectExtension(t ObjectDesc, impl interface{}) {
	cfg := t.Config()
	fields := cfg.Fields.(graphql.Fields)
	for fieldName, handler := range t.FieldHandlers {
		fields[fieldName].Resolve = handler(impl)
	}
	service.types.addExtension(cfg.Name, cfg)
}

// RegisterUnion registers a GraphQL type with the service.
func (service *Service) RegisterUnion(t UnionDesc, impl UnionTypeResolver) {
	cfg := t.Config()
	registrar := func(m graphql.TypeMap) graphql.Type {
		cfg = t.Config()
		newTypes := make([]*graphql.Object, len(cfg.Types))
		for i, t := range cfg.Types {
			objType := m[t.PrivateName].(*graphql.Object)
			newTypes[i] = objType
		}
		cfg.Types = newTypes

		cfg.ResolveType = nil
		if impl != nil {
			cfg.ResolveType = newResolveTypeFn(m, impl)
		}
		return graphql.NewUnion(cfg)
	}
	service.types.addType(cfg.Name, UnionKind, registrar)
}

// RegisterSchema registers given GraphQL schema with the service.
func (service *Service) RegisterSchema(t SchemaDesc) {
	service.types.setSchema(t)
}

// RegisterMiddleware registers given middleware with the service.
func (service *Service) RegisterMiddleware(mware Middleware) {
	service.mware = append(service.mware, mware)
}

// Regenerate generates new schema given registered types.
func (service *Service) Regenerate() error {
	schema, err := newSchema(service.types, service.mware)
	if err == nil {
		service.schema = schema
	}
	return err
}

// Do executes request given query string
func (service *Service) Do(ctx context.Context, q string, vars map[string]interface{}) *graphql.Result {
	schema := service.schema
	params := graphql.Params{
		Schema:         schema,
		VariableValues: vars,
		Context:        ctx,
		RequestString:  q,
	}

	// run init middleware
	MiddlewareHandleInits(service, &params)

	// parse the source
	parseFinishFn := MiddlewareHandleParseDidStart(service, &params)
	source := source.NewSource(&source.Source{
		Body: []byte(q),
		Name: "GraphQL request",
	})
	AST, err := parser.Parse(parser.ParseParams{Source: source})
	parseFinishFn(err)
	if err != nil {
		return &graphql.Result{Errors: gqlerrors.FormatErrors(err)}
	}

	// validate document
	validationFinishFn := MiddlewareHandleValidationDidStart(service, &params)
	validationResult := graphql.ValidateDocument(&schema, AST, nil)
	validationFinishFn(validationResult.Errors)
	if !validationResult.IsValid {
		return &graphql.Result{Errors: validationResult.Errors}
	}

	// execute query
	return service.Executor(graphql.ExecuteParams{
		Schema:  schema,
		AST:     AST,
		Args:    vars,
		Context: ctx,
	})
}

type typeRegister struct {
	types      map[Kind]map[string]registerTypeFn
	extensions map[string][]interface{}
	schema     SchemaDesc
}

func newTypeRegister() *typeRegister {
	exts := map[string][]interface{}{}
	types := make(map[Kind]map[string]registerTypeFn, 6)
	types[EnumKind] = map[string]registerTypeFn{}
	types[ScalarKind] = map[string]registerTypeFn{}
	types[ObjectKind] = map[string]registerTypeFn{}
	types[InputKind] = map[string]registerTypeFn{}
	types[InterfaceKind] = map[string]registerTypeFn{}
	types[UnionKind] = map[string]registerTypeFn{}
	return &typeRegister{types: types, extensions: exts}
}

func (r *typeRegister) addType(name string, kind Kind, fn registerTypeFn) {
	if r.types == nil {
		r.types = map[Kind]map[string]registerTypeFn{}
	}
	if _, ok := r.types[kind]; !ok {
		r.types[kind] = map[string]registerTypeFn{}
	}
	r.types[kind][name] = fn
}

func (r *typeRegister) addExtension(name string, cfg interface{}) {
	if _, ok := r.extensions[name]; !ok {
		r.extensions[name] = []interface{}{}
	}
	r.extensions[name] = append(r.extensions[name], cfg)
}

func (r *typeRegister) extensionsForType(t string) []interface{} {
	if _, ok := r.extensions[t]; !ok {
		return []interface{}{}
	}
	return r.extensions[t]
}

func (r *typeRegister) setSchema(desc SchemaDesc) {
	r.schema = desc
}

func newSchema(reg *typeRegister, mware []Middleware) (graphql.Schema, error) {
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
		queryType := findType(typeMap, schemaCfg.Query.Name())
		schemaCfg.Query = queryType.(*graphql.Object)
	}
	if schemaCfg.Mutation != nil {
		mutationType := findType(typeMap, schemaCfg.Mutation.Name())
		schemaCfg.Mutation = mutationType.(*graphql.Object)
	}
	if schemaCfg.Subscription != nil {
		subscriptionType := findType(typeMap, schemaCfg.Subscription.Name())
		schemaCfg.Subscription = subscriptionType.(*graphql.Object)
	}

	schema, err := graphql.NewSchema(schemaCfg)
	if err != nil {
		return schema, err
	}

	// Types that are not directly referenced by the root Schema type or any of
	// their children are not immediately registered with the schema. As such to
	// ensure that ALL types are available we append any that are missing.
	registeredTypes := schema.TypeMap()
	for _, ltype := range typeMap {
		if _, registered := registeredTypes[ltype.Name()]; registered {
			continue
		}
		if err = schema.AppendType(ltype); err != nil {
			return schema, err
		}
	}

	// Register middleware
	for _, m := range mware {
		schema.AddExtensions(m)
	}

	return schema, err
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

func findType(m graphql.TypeMap, name string) graphql.Type {
	if t, ok := m[name]; ok {
		return t
	}
	panic(
		fmt.Sprintf("required type '%s' not registered.", name),
	)
}

func mergeObjectConfig(a, b graphql.ObjectConfig) {
	af := a.Fields.(graphql.Fields)
	bf := b.Fields.(graphql.Fields)
	for n, f := range bf {
		af[n] = f
	}
	ai := a.Interfaces.([]*graphql.Interface)
	bi := a.Interfaces.([]*graphql.Interface)
	a.Interfaces = append(ai, bi...)
}
