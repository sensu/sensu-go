package generator

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genObjectType(node *ast.ObjectDefinition) jen.Code {
	code := newGroup()
	name := node.GetName().Value
	resolverName := fmt.Sprintf("%sResolver", name)

	//
	// Generate resolver interface
	//
	// ... comment: Describe resolver interface and usage
	// ... method:  [one method for each field]
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
//   type DogResolver interface {
//     // Name implements response to request for name field.
//     Name(graphql.Params) interface{}
//     // Breed implements response to request for breed field.
//     Breed(graphql.Params) interface{}
//     // IsTypeOf is used to determine if a given value is associated with the Dog type
//     IsTypeOf(graphql.IsTypeOfParams) bool
//   }
//
// == Example implementation ...
//
//   // MyDogResolver implements DogResolver interface
//   type MyDogResolver struct {
//     logger logrus.LogEntry
//     store interface{
//       store.BreedStore
//       store.DogStore
//     }
//   }
//
//   // Name implements response to request for name field.
//   func (r *MyDogResolver) Name(p graphql.Params) (interface{}, error) {
//     // ... implementation details ...
//     dog := p.Source.(DogGetter)
//     return dog.GetName()
//   }
//
//   // Breed implements response to request for breed field.
//   func (r *MyDogResolver) Name(p graphql.Params) (interface{}, error) {
//     // ... implementation details ...
//     dog := p.Source.(DogGetter)
//     breed := r.store.GetBreed(dog.GetBreedName())
//     return breed
//   }
//
//   // IsTypeOf is used to determine if a given value is associated with the Dog type
//   func (r *MyDogResolver) IsTypeOf(p graphql.IsTypeOfParams) bool {
//     // ... implementation details ...
//     _, ok := p.Value.(DogGetter)
//     return ok
//   }`,
		resolverName,
		name,
	)
	// Generate resolver interface.
	code.Type().Id(resolverName).InterfaceFunc(func(g *jen.Group) {
		for _, field := range node.Fields {
			// Define method for each field in object type
			name := field.Name.Value
			titleizedName := toFieldName(field.Name.Value)

			// func FieldName(params graphql.Params) (interface{}, error)
			g.Commentf("%s implements response to request for '%s' field.", titleizedName, name)
			g.Id(titleizedName).Params(
				jen.Qual(graphqlGoPkg, "ResolveParams"),
			).Parens(jen.List(jen.Interface(), jen.Error()))
		}

		// Satisfy IsTypeOf() callback
		g.Commentf("IsTypeOf is used to determine if a given value is associated with the %s type", name)
		g.Id("IsTypeOf").Params( // IsTypeOf(graphql.IsTypeOfParams) bool
			jen.Qual(graphqlGoPkg, "IsTypeOfParams"),
		).Bool()
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
//   }`,
		aliasResolver,
		resolverName,
	)
	code.Type().Id(aliasResolver).Struct()
	for _, field := range node.Fields {
		// Define method for each field in object type
		name := field.Name.Value
		titleizedName := toFieldName(field.Name.Value)

		code.Commentf("%s implements response to request for '%s' field.", titleizedName, name)
		code.Func().Params(jen.Id("_").Id(aliasResolver)).Id(titleizedName).Params(
			jen.Id("p").Qual(graphqlGoPkg, "ResolveParams"),
		).Block(jen.Return(jen.Qual(servicePkg, "DefaultResolver").Call(
			jen.Id("p").Dot("Source"),
			jen.Id("p").Dot("Info").Dot("FieldName"),
		)))
	}

	//
	// Generate type definition
	//
	// ... comment: Include description in comment
	// ... panic callbacks panic if not configured
	//

	// Object ype description
	typeDesc := fetchDescription(node)

	// To appease the linter ensure that the the description of the object type
	// begins with the name of the resulting method.
	desc := typeDesc
	if hasPrefix := strings.HasPrefix(typeDesc, name); !hasPrefix {
		desc = name + " " + desc
	}

	// Generate interface references
	ints := jen.Index().Op("*").Qual(graphqlGoPkg, "Interface").ValuesFunc(
		func(g *jen.Group) {
			for _, n := range node.Interfaces {
				g.Line().Add(genMockInterfaceReference(n))
			}
		},
	)

	privateNamePrefix := "_ObjType_" + name

	//
	// Generates thunk that returns new instance of object config
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
	//   // Dogs are not hooman
	//   var Dog = graphql.NewType("Dog", graphql.ObjectKind)
	//
	//   // RegisterDog registers Dog type with given service
	//   func RegisterDog(scv graphql.Service, impl DogResolver) {
	//     return scv.RegisterObject(_ObjType_Dog_Desc, impl)
	//   }
	//
	//   // DogNameResolverParams describes args and context given to field resolver.
	//   type DogNameResolverParams struct {
	//     graphql.ResolveParams
	//     Args DogNameResolverArgs
	//   }
	//
	//   // DogNameResolverArgs describes user arguments given when selecting field.
	//   type DogNameResolverArgs struct {
	//     Style string
	//   }
	//
	//   func _ObjType_Dog_Name(impl interface{}, p desc.ResolveParams) (inteface{}, error) {
	//     args := DogNameResolverArgs{}
	//     err := mapstructure.Decode(p.Args, args)
	//     if err != nil {
	//       return err
	//     }
	//     params := DogNameResolverParams{ResolveParams: p, Args: args}
	//     return impl.(DogResolver).Name(params)
	//   }
	//
	//   func _ObjType_Dog_Breed(impl interface{}, p desc.ResolveParams) (inteface{}, error) {
	//     return impl.(DogResolver).Breed(p)
	//   }
	//
	//   func Dog() graphql.ObjectConfig { // implements TypeThunk
	//     return graphql.ObjectConfig{
	//       Name:        "Dog",
	//       Description: "Dogs are not hooman",
	//       Interfaces:  // ...
	//       Fields:      // ...
	//       IsKindOf:    // ...
	//     }
	//   }
	//
	code.Comment(desc)
	code.Id(name).Op("=").Id(privateNamePrefix)
	code.Id(name).Struct()
	code.Func().Id(name).Params().Qual(graphqlGoPkg, "ObjectConfig").Block(
		jen.Return(jen.Qual(graphqlGoPkg, "ObjectConfig").Values(jen.Dict{
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(typeDesc),
			jen.Id("Interfaces"):  ints,
			jen.Id("Fields"):      genFields(node.Fields),
			jen.Id("IsTypeOf"): jen.Func().Params(jen.Id("_").Qual(graphqlGoPkg, "IsTypeOfParams")).Bool().Block(
				jen.Comment(missingResolverNote),
				jen.Panic(jen.Lit("Unimplemented; see "+resolverName+".")),
			),
		})),
	)

	return code
}
