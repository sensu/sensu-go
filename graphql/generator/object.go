package generator

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

//
// Generate config for enum
//
// == Example input SDL
//
//   """
//   Dogs are not hooman.
//   """
//   type Dog implements Pet {
//     "name of this fine beast."
//     name(style: String = "full"):  String!
//
//     "breed of this silly animal; probably shibe."
//     breed: [Breed]
//   }
//
// == Example output
//
//   // DogNameFieldResolver ...
//   type DogNameFieldResolver interface {
//     // Name implements response to request for name field.
//     Name(graphql.Params) interface{}
//   }
//
//   // DogBreedFieldResolver ...
//   type DogBreedFieldResolver interface {
//     // Breed implements response to request for breed field.
//     Breed(graphql.Params) interface{}
//   }
//
//   // DogFieldResolvers ...
//   type DogFieldResolvers interface {
//     DogNameFieldResolver
//     DogBreedFieldResolver
//
//     // IsTypeOf is used to determine if a given value is associated with the Dog type
//     IsTypeOf(interface{}, graphql.IsTypeOfParams) bool
//   }
//
//   // DogAliases ...
//   type DogAliases struct {}
//
//   // Name ...
//   func (_ DogAliases) Name(p graphql.ResolveParams) (interface{}, error) {
//     return graphql.DefaultResolver(p.Source, p.Info.FieldName)
//   }
//
//   // Breed ...
//   func (_ DogAliases) Breed(p graphql.ResolveParams) (interface{}, error) {
//     return graphql.DefaultResolver(p.Source, p.Info.FieldName)
//   }
//
//   // DogType ...
//   var DogType = graphql.NewType("Dog", graphql.ObjectKind)
//
//   // RegisterDog registers Dog object type with given service.
//   func RegisterDog(svc graphql.Service, impl DogFieldResolvers) {
//     svc.RegisterObect(_ObjectTypeDog, impl)
//   }
//
func genObjectType(node *ast.ObjectDefinition, i info) jen.Code {
	code := newGroup()
	name := node.GetName().Value
	desc := getDescription(node)

	// Ids ...
	fieldResolversName := name + "FieldResolvers"
	publicRefName := name + "Type"
	publicRefComment := genTypeComment(publicRefName, desc)
	privateConfigName := mkPrivateID(node, "Desc")
	privateConfigThunkName := mkPrivateID(node, "ConfigFn")

	//
	// Generate field resolver interfaces
	//
	//
	// == Example output
	//
	//   // DogNameFieldResolver ...
	//   type DogNameFieldResolver interface {
	//     // Name implements response to request for name field.
	//     Name(graphql.ResolveParams) (interface{}, error)
	//   }
	//
	for _, f := range node.Fields {
		resolverCode := genFieldResolverInterface(f, i)
		code.Add(resolverCode)
		code.Line()
	}

	//
	// Generate resolver interface
	//
	// == Example output
	//
	//   // DogFieldResolvers ...
	//   type DogFieldResolvers interface {
	//     DogNameFieldResolver
	//     DogBreedFieldResolver
	//
	//     // IsTypeOf is used to determine if a given value is associated with the Dog type
	//     IsTypeOf(interface{}, graphql.IsTypeOfParams) bool
	//   }
	//
	code.Commentf(`//
// %s represents a collection of methods whose products represent the
// response values of the '%s' type.
//
// == Example SDL
//
//   """
//   Dog's are not hooman.
//   """
//   type Dog implements Pet {
//     "name of this fine beast."
//     name:  String!
//
//     "breed of this silly animal; probably shibe."
//     breed: [Breed]
//   }
//
// == Example generated interface
//
//   // DogResolver ...
//   type DogFieldResolvers interface {
//     DogNameFieldResolver
//     DogBreedFieldResolver
//
//     // IsTypeOf is used to determine if a given value is associated with the Dog type
//     IsTypeOf(interface{}, graphql.IsTypeOfParams) bool
//   }
//
// == Example implementation ...
//
//   // DogResolver implements DogFieldResolvers interface
//   type DogResolver struct {
//     logger logrus.LogEntry
//     store interface{
//       store.BreedStore
//       store.DogStore
//     }
//   }
//
//   // Name implements response to request for name field.
//   func (r *DogResolver) Name(p graphql.ResolveParams) (interface{}, error) {
//     // ... implementation details ...
//     dog := p.Source.(DogGetter)
//     return dog.GetName()
//   }
//
//   // Breed implements response to request for breed field.
//   func (r *DogResolver) Breed(p graphql.ResolveParams) (interface{}, error) {
//     // ... implementation details ...
//     dog := p.Source.(DogGetter)
//     breed := r.store.GetBreed(dog.GetBreedName())
//     return breed
//   }
//
//   // IsTypeOf is used to determine if a given value is associated with the Dog type
//   func (r *DogResolver) IsTypeOf(p graphql.IsTypeOfParams) bool {
//     // ... implementation details ...
//     _, ok := p.Value.(DogGetter)
//     return ok
//   }
//`,
		fieldResolversName,
		name,
	)
	// Generate resolver interface.
	code.
		Type().Id(fieldResolversName).
		InterfaceFunc(func(g *jen.Group) {
			// Include each field resolver
			for _, field := range node.Fields {
				resolverName := genFieldResolverName(field, i)
				g.Id(resolverName)
			}
			g.Line()

			// TODO: Removed until I decide what I want to do with it.
			// Satisfy IsTypeOf() callback
			// IsTypeOf(graphql.IsTypeOfParams) bool
			// g.Commentf(
			// 	"IsTypeOf is used to determine if a given value is associated with the %s type",
			// 	name,
			// )
			// g.Id("IsTypeOf").Params(
			// 	jen.Interface(),
			// 	jen.Qual(servicePkg, "IsTypeOfParams"),
			// ).Bool()
		})

	//
	// Generate alias implementation of resolver interface
	//
	// ... comment:   Include description of usage
	// ... method(s): Implement all methods
	//

	aliasResolver := fmt.Sprintf("%sAliases", name)
	code.Commentf(`// %s implements all methods on %s interface by using reflection to
// match name of field to a field on the given value. Intent is reduce friction
// of writing new resolvers by removing all the instances where you would simply
// have the resolvers method return a field.
//
// == Example SDL
//
//    type Dog {
//      name:   String!
//      weight: Float!
//      dob:    DateTime
//      breed:  [Breed]
//    }
//
// == Example generated aliases
//
//   type DogAliases struct {}
//   func (_ DogAliases) Name(p graphql.ResolveParams) (interface{}, error) {
//     // reflect...
//   }
//   func (_ DogAliases) Weight(p graphql.ResolveParams) (interface{}, error) {
//     // reflect...
//   }
//   func (_ DogAliases) Dob(p graphql.ResolveParams) (interface{}, error) {
//     // reflect...
//   }
//   func (_ DogAliases) Breed(p graphql.ResolveParams) (interface{}, error) {
//     // reflect...
//   }
//
// == Example Implementation
//
//   type DogResolver struct { // Implements DogResolver
//     DogAliases
//     store store.BreedStore
//   }
//
//   // NOTE:
//   // All other fields are satisified by DogAliases but since this one
//   // requires hitting the store we implement it in our resolver.
//   func (r *DogResolver) Breed(p graphql.ResolveParams) interface{} {
//     dog := v.(*Dog)
//     return r.BreedsById(dog.BreedIDs)
//   }
//`,
		aliasResolver,
		fieldResolversName,
	)
	code.Type().Id(aliasResolver).Struct()
	for _, field := range node.Fields {
		// Define method for each field in object type
		name := field.Name.Value
		titleizedName := toFieldName(field.Name.Value)
		resolverFnSignature := genFieldResolverSignature(field, i)
		coerceType := genFieldResolverTypeCoercion(field.Type, i)

		code.Commentf("%s implements response to request for '%s' field.", titleizedName, name)
		code.
			Func().Params(jen.Id("_").Id(aliasResolver)).
			Add(resolverFnSignature).
			BlockFunc(func(g *jen.Group) {
				g.List(jen.List(jen.Id("val"), jen.Id("err"))).Op(":=").
					Qual(servicePkg, "DefaultResolver").
					Call(
						jen.Id("p").Dot("Source"),
						jen.Id("p").Dot("Info").Dot("FieldName"),
					)
				if coerceType != nil {
					g.Id("ret").Op(":=").Add(coerceType) //Id("val").Assert(retType)
					g.Return(jen.List(jen.Id("ret"), jen.Id("err")))
				} else {
					g.Return(jen.List(jen.Id("val"), jen.Id("err")))
				}
			})
	}

	//
	// Generate public reference to type
	//
	// == Example output
	//
	//   // DogType ... Dogs are great!
	//   var DogType = graphql.NewType("Dog", graphql.ObjectKind)
	//
	code.Comment(publicRefComment)
	code.
		Var().Id(publicRefName).Op("=").
		Qual(servicePkg, "NewType").
		Call(jen.Lit(name), jen.Qual(servicePkg, "ObjectKind"))

	//
	// Generate public func to register type with service
	//
	// == Example output
	//
	//   // RegisterDog registers Dog object type with given service.
	//   func RegisterDog(svc graphql.Service, impl DogFieldResolvers) {
	//     svc.RegisterObject(_ObjTypeDogDesc, impl)
	//   }
	//

	code.Add(
		genRegisterFn(node, jen.Id(fieldResolversName)),
	)

	//
	// Generate field handlers
	//
	// == Example output
	//
	//   func _ObjTypeDogNameHandler(impl interface{}) graphql.FieldResolveFn {
	//     resolver := impl.(DogNameFieldResolver)
	//     return func(p graphql.ResolveParams) (interface{}, error) {
	//       arg := DogNameResolverArgs{}
	//       err := mapstructure.Decode(p.Args, args)
	//       if err != nil {
	//         return err
	//       }
	//
	//       params := DogNameResolverParams{ResolveParams: p, Args: arg}
	//       return resolver.Name(params)
	//     }
	//   }
	//
	//   func _ObjTypeDogBreedHandler(impl interface{}) graphql.FieldResolveFn {
	//     resolver := impl.(DogBreedFieldResolver)
	//     return resolver.Breed
	//   }
	//
	for _, f := range node.Fields {
		handler := genFieldHandlerFn(f, i)
		code.Add(handler)
		code.Line()
	}

	//
	// Generates thunk that returns new instance of object config
	//
	//   func _ObjTypeDogConfigFn() graphql.ObjectConfig {
	//     return graphql.ObjectConfig{
	//       Name:        "Dog",
	//       Description: "Dogs are not hooman",
	//       Interfaces:  // ...
	//       Fields:      // ...
	//       IsKindOf:    // ...
	//     }
	//   }
	//

	// Generate interface references
	ints := jen.
		Index().Op("*").Qual(defsPkg, "Interface").
		ValuesFunc(
			func(g *jen.Group) {
				for _, n := range node.Interfaces {
					g.Line().Add(genMockInterfaceReference(n))
				}
			},
		)

	// Generaate default IsTypeOfParams handler
	typeOfFn := jen.
		Func().Params(jen.Id("_").
		Qual(defsPkg, "IsTypeOfParams")).Bool().
		Block(
			jen.Comment(missingResolverNote),
			jen.Panic(jen.Lit("Unimplemented; see "+fieldResolversName+".")),
		)

	// Generate config thunk
	code.
		Func().Id(privateConfigThunkName).
		Params().Qual(defsPkg, "ObjectConfig").
		Block(
			jen.Return(jen.Qual(defsPkg, "ObjectConfig").Values(jen.Dict{
				jen.Id("Name"):        jen.Lit(name),
				jen.Id("Description"): jen.Lit(desc),
				jen.Id("Interfaces"):  ints,
				jen.Id("Fields"):      genFields(node.Fields),
				jen.Id("IsTypeOf"):    typeOfFn,
			})),
		)

	//
	// Generate type description
	//
	// == Example output
	//
	//   // describe dog's configuration; kept private to avoid unintentional
	//   // tampering at runtime.
	//   var _ObjTypeDogDesc = graphql.ObjectConfig{
	//     Config: _ObjTypeDogConfigFn,
	//     FieldHandlers: map[string]graphql.FieldHandler{
	//       "id":    _ObjTypeDogIDHandler,
	//       "name":  _ObjTypeDogNameHandler,
	//       "breed": _ObjTypeDogBreedHandler,
	//     }
	//   }
	//
	code.Commentf(
		`describe %s's configuration; kept private to avoid unintentional tampering of configuration at runtime.`,
		name,
	)
	code.
		Var().Id(privateConfigName).Op("=").
		Qual(servicePkg, "ObjectDesc").
		Values(jen.Dict{
			jen.Id("Config"): jen.Id(privateConfigThunkName),
			jen.Id("FieldHandlers"): jen.Map(jen.String()).Qual(servicePkg, "FieldHandler").Values(jen.DictFunc(func(d jen.Dict) {
				for _, f := range node.Fields {
					key := f.Name.Value
					handlerName := genFieldHandlerName(f, i)
					d[jen.Lit(key)] = jen.Id(handlerName)
				}
			})),
		})

	return code
}

//
// Generate field resolver interface for given field
//
// == Example input SDL
//
//   """
//   Dogs are not hooman.
//   """
//   type Dog {
//     "name of this fine beast."
//     name: String!
//   }
//
// == Example output
//
//   // DogFieldResolvers ...
//   type DogNameFieldResolvers interface {
//     // Name implements response to request for name field.
//     Name(graphql.ResolveParams) (string, error)
//   }
//
// == Example input SDL
//
//   """
//   Dogs are not hooman.
//   """
//   type Dog {
//     "name of this fine beast."
//     name: NameComponents
//   }
//
// == Example output
//
//   // DogNameFieldResolver ...
//   type DogNameFieldResolver interface {
//     // Name implements response to request for name field.
//     Name(graphql.ResolveParams) (interface{}, error)
//   }
//
// == Example input SDL
//
//   """
//   Dogs are not hooman.
//   """
//   type Dog {
//     "name of this fine beast."
//     name(style: String = "full", locale: Locale = EN): String!
//   }
//
// == Example output
//
//   // DogNameFieldArgs ...
//   type DogNameFieldArgs struct {
//     Style string
//     Locale Locale
//   }
//
//   // DogNameFieldParams ...
//   type DogNameFieldParams struct {
//     graphql.ResolveParams
//     Args DogNameFieldArgs
//   }
//
//   // DogNameFieldResolver ...
//   type DogNameFieldResolver interface {
//     // Name implements response to request for name field.
//     Name(DogNameFieldParams) (interface{}, error)
//   }
//
// == Example input SDL
//
//   type Mutation {
//     // updateAvatar updates given doggo's profile picture.
//     updateAvatar(inputs: UpdateAvatarInput!): UpdateAvatarPayload!
//   }
//
// == Example output
//
//   // MutationUpdateAvatarFieldArgs ...
//   type MutationUpdateAvatarFieldArgs struct {
//     Inputs *UpdateAvatarInput
//   }
//
//   // MutationUpdateAvatarFieldParams ...
//   type MutationUpdateAvatarFieldParams struct {
//     graphql.ResolveParams
//     Args MutationUpdateAvatarFieldArgs
//   }
//
//   // MutationUpdateAvatarFieldResolver ...
//   type MutationUpdateAvatarFieldResolver interface {
//     // Name implements response to request for name field.
//     UpdateAvatar(MutationUpdateAvatarFieldParams) (interface{}, error)
//   }
//
func genFieldResolverInterface(field *ast.FieldDefinition, i info) jen.Code {
	code := newGroup()

	// names
	typeName := i.currentNode
	fieldName := field.Name.Value

	//
	// If field has arguments create type to encapsulate parameters.
	//
	// == Example output
	//
	//   // DogNameFieldArgs ...
	//   type DogNameFieldArgs struct {
	//     Style string
	//     Locale Locale
	//   }
	//
	if len(field.Arguments) > 0 {
		argsName := genFieldResolverArgsName(field, i)
		code.Commentf("%s contains arguments provided to %s when selected", argsName, fieldName)
		code.Type().Id(argsName).StructFunc(func(g *jen.Group) {
			for _, arg := range field.Arguments {
				retType := genConcreteTypeReference(arg.Type, i)
				fieldName := toFieldName(arg.Name.Value)
				comment := genFieldComment(fieldName, getDescription(arg), "")
				g.Id(fieldName).Add(retType).Comment(comment)
			}
		})

		paramsName := genFieldResolverParamsName(field, i)
		code.Commentf("%s contains contextual info to resolve %s field", paramsName, fieldName)
		code.Type().Id(paramsName).Struct(
			jen.Qual(servicePkg, "ResolveParams"),
			jen.Id("Args").Id(argsName),
		)
	}

	//
	// Generate field resolver interface
	//
	// == Example output
	//
	//   // DogNameFieldResolver ...
	//   type DogNameFieldResolver interface {
	//     // Name implements response to request for name field.
	//     Name(DogNameFieldParams) (interface{}, error)
	//   }
	//
	resolverName := genFieldResolverName(field, i)
	resolverFnName := toFieldName(fieldName)
	resolverFnSignature := genFieldResolverSignature(field, i)
	code.Commentf(
		"%s implement to resolve requests for the %s's %s field.",
		resolverName,
		typeName,
		fieldName,
	)
	code.Type().Id(resolverName).Interface(
		jen.Commentf(
			"%s implements response to request for %s field.",
			resolverFnName,
			fieldName,
		),
		resolverFnSignature,
	)

	return code
}

//
// == Examples
//
//   breed: String       => BreedFieldResolver
//   id: String!         => IDFieldResolver
//   profilePicture: URL => ProfilePictureFieldResolver
//
func genFieldResolverName(field *ast.FieldDefinition, i info) string {
	typeName := strings.Title(i.currentNode)
	fieldName := toFieldName(field.Name.Value)
	return typeName + fieldName + "FieldResolver"
}

//
// == Examples
//
//   breed: String       => BreedFieldResolverArgs
//   id: String!         => IDFieldResolverArgs
//   profilePicture: URL => ProfilePictureFieldResolverArgs
//
func genFieldResolverArgsName(field *ast.FieldDefinition, i info) string {
	return genFieldResolverName(field, i) + "Args"
}

//
// == Examples
//
//   breed: String       => BreedFieldResolverParams
//   id: String!         => IDFieldResolverParams
//   profilePicture: URL => ProfilePictureFieldResolverParams
//
func genFieldResolverParamsName(field *ast.FieldDefinition, i info) string {
	return genFieldResolverName(field, i) + "Params"
}

//
// == Examples
//
//   breed: String       => Breed(BreedFieldResolverParams) (interface{}, error)
//   id: String!         => ID(IDFieldResolverParams) (interface{}, error)
//   profilePicture: URL => ProfilePicture(ProfilePictureFieldResolverParams) (interface, error)
//
func genFieldResolverSignature(field *ast.FieldDefinition, i info) jen.Code {
	// method name
	fieldName := toFieldName(field.Name.Value)

	// parameters
	params := jen.Id("p").Qual(servicePkg, "ResolveParams")
	if len(field.Arguments) > 0 {
		params = jen.Id("p").Id(genFieldResolverParamsName(field, i))
	}

	// return type
	retType := genFieldResolverReturnType(field.Type, i)
	return jen.Id(fieldName).Params(params).Parens(jen.List(retType, jen.Error()))
}

//
// == Examples
//
//   GraphQL => Go
//
//   String  => string
//   [String]=> []string
//   [Int]   => []int
//   [Int!]  => []int
//   [Int!]! => []int
//   Int     => int
//   Int!    => int
//   Bool    => bool
//   MyObj   => interface{}
//   [MyObj] => interface{}
//   MyObj!  => interface{}
//
func genFieldResolverReturnType(t ast.Type, i info) jen.Code {
	var namedType *ast.Named
	switch ttype := t.(type) {
	case *ast.List:
		// Super crufty.
		var ok bool
		namedType, ok = ttype.Type.(*ast.Named) // ok is true if type is [String]
		if !ok {
			nullType, ok := ttype.Type.(*ast.NonNull) // ok is true if type is [String!]
			if !ok {
				return jen.Interface()
			}
			namedType, ok = nullType.Type.(*ast.Named) // is is true if type isn't list
			if !ok {
				return jen.Interface()
			}
		}
		statement := genBuiltinTypeReference(namedType)
		if statement != nil {
			return jen.Index().Add(statement)
		}
		return jen.Interface()
	case *ast.NonNull:
		return genFieldResolverReturnType(ttype.Type, i)
	case *ast.Named:
		namedType = ttype
	default:
		panic("unknown ast.Type given")
	}

	// Check if type matches definition
	if def, ok := i.definitions[namedType.Name.Value]; ok {
		// Use type if enum
		if _, ok := def.(*ast.EnumDefinition); ok {
			return jen.Id(namedType.Name.Value)
		}

		// Otherwise simply fallback to interface{}
		return jen.Interface()
	}

	// Handle as built-in type if it doesn't match any user defined type.
	if code := genBuiltinTypeReference(namedType); code != nil {
		return code
	}
	return jen.Interface()
}

// Super crufty.
func genFieldResolverTypeCoercion(t ast.Type, i info) jen.Code {
	mkAssert := func(t jen.Code) jen.Code {
		return jen.Id("val").Assert(t)
	}

	var namedType *ast.Named
	switch ttype := t.(type) {
	case *ast.List:
		var ok bool
		namedType, ok = ttype.Type.(*ast.Named)
		if !ok {
			nullType, ok := ttype.Type.(*ast.NonNull)
			if !ok {
				return nil
			}
			namedType, ok = nullType.Type.(*ast.Named)
			if !ok {
				return nil
			}
		}
		statement := genBuiltinTypeReference(namedType)
		if statement != nil {
			return mkAssert(jen.Index().Add(statement))
		}
		return nil
	case *ast.NonNull:
		return genFieldResolverTypeCoercion(ttype.Type, i)
	case *ast.Named:
		namedType = ttype
	default:
		panic("unknown ast.Type given")
	}

	if def, ok := i.definitions[namedType.Name.Value]; ok {
		if _, ok := def.(*ast.EnumDefinition); ok {
			return jen.Id(namedType.Name.Value).Call(mkAssert(jen.String()))
		}
		return nil
	}

	switch namedType.Name.Value {
	case graphql.Int.Name():
		return jen.Qual(defsPkg, "Int").Op(".").Id("ParseValue").Call(jen.Id("val")).Assert(jen.Int())
	case graphql.Float.Name():
		return jen.Qual(defsPkg, "Float").Op(".").Id("ParseValue").Call(jen.Id("val")).Assert(jen.Float64())
	case graphql.String.Name():
		return jen.Qual("fmt", "Sprint").Call(jen.Id("val"))
	case graphql.Boolean.Name():
		return mkAssert(jen.Id("bool"))
	case graphql.DateTime.Name():
		return mkAssert(jen.Qual("time", "Time"))
	}
	return nil
}

//
// == Examples
//
//   breed: String       => _ObjTypeDogBreedHandler
//   id: String!         => _ObjTypeDogIDHandler
//   profilePicture: URL => _ObjTypeDogProfilePictureHandler
//
func genFieldHandlerName(field *ast.FieldDefinition, i info) string {
	typeName := strings.Title(i.currentNode)
	fieldName := toFieldName(field.Name.Value)
	return "_ObjType" + typeName + fieldName + "Handler"
}

//
// Generate field handlers
//
// == Example SDL
//
//   type Dog {
//     name(style: String = "full", locale: Locale = EN): String!
//     breed: [String]
//   }
//
// == Example output
//
//   func _ObjTypeDogNameHandler(impl interface{}) graphql.FieldResolveFn {
//     resolver := impl.(DogNameFieldResolver)
//     return func(p graphql.ResolveParams) (interface{}, error) {
//       frp := DogNameResolverParams{ResolveParams: p}
//       err := mapstructure.Decode(p.Args, frp.Args)
//       if err != nil {
//         return nil, err
//       }
//
//       return resolver.Name(frp)
//     }
//   }
//
//   func _ObjTypeDogBreedHandler(impl interface{}) graphql.FieldResolveFn {
//     resolver := impl.(DogBreedFieldResolver)
//     return resolver.Breed
//   }
//
func genFieldHandlerFn(field *ast.FieldDefinition, i info) jen.Code {
	fieldName := toFieldName(field.Name.Value)
	handlerName := genFieldHandlerName(field, i)

	return jen.
		Func().Id(handlerName).
		Params(jen.Id("impl").Interface()).
		Qual(defsPkg, "FieldResolveFn").
		BlockFunc(func(g *jen.Group) {
			// eg. resolver := impl.(DogNameFieldResolver)
			fieldResolverName := genFieldResolverName(field, i)
			g.Id("resolver").Op(":=").Id("impl").Assert(jen.Id(fieldResolverName))

			// If field has arguments, use generated parameters type
			if len(field.Arguments) > 0 {
				fieldResolverParamsName := genFieldResolverParamsName(field, i)

				g.Return(
					jen.Func().
						Params(jen.Id("p").Qual(defsPkg, "ResolveParams")).
						Parens(jen.List(jen.Interface(), jen.Error())).
						Block(
							jen.Id("frp").Op(":=").Id(fieldResolverParamsName).Values(jen.Dict{
								jen.Id("ResolveParams"): jen.Id("p"),
							}),
							jen.Id("err").Op(":=").Qual(mapstructurePkg, "Decode").Call(
								jen.Id("p.Args"),
								jen.Op("&").Id("frp.Args"),
							),
							jen.If(jen.Id("err").Op("!=").Nil()).Block(
								jen.Return(jen.List(jen.Nil(), jen.Id("err"))),
							),
							jen.Line(),
							jen.Return(jen.Id("resolver."+fieldName).Call(jen.Id("frp"))),
						),
				)
			} else {
				g.Return(
					jen.Func().
						Params(jen.Id("p").Qual(defsPkg, "ResolveParams")).
						Parens(jen.List(jen.Interface(), jen.Error())).
						Block(
							jen.Return(
								jen.Id("resolver." + fieldName).Call(jen.Id("p")),
							),
						),
				)
			}
		})
}
