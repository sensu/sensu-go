package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/graphql-go/graphql/language/ast"
)

func genFieldResolversInterface(name string, node *ast.ObjectDefinition, i info) jen.Code {
	interfaceName := mkFieldResolversName(name)

	code := newGroup()
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
		interfaceName,
		name,
	)
	// Generate resolver interface.
	code.
		Type().Id(interfaceName).
		InterfaceFunc(func(g *jen.Group) {
			// Include each field resolver
			for _, field := range node.Fields {
				resolverName := genFieldResolverName(field, i)
				g.Id(resolverName)
			}
			g.Line()
		})
	return code
}

func genFieldAliases(name string, node *ast.ObjectDefinition, i info) jen.Code {
	fieldResolversName := mkFieldResolversName(name)
	aliasResolver := fmt.Sprintf("%sAliases", name)

	code := newGroup()
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
		coerceType := genFieldResolverTypeCoercion(field.Type, i, true)

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
					g.List(jen.Id("ret"), jen.Id("ok")).Op(":=").Add(coerceType)
					g.If(jen.Id("err").Op("!=").Nil()).Block(
						jen.Return(jen.List(jen.Id("ret"), jen.Id("err"))),
					)
					g.If(jen.Op("!").Id("ok")).Block(
						jen.Return(jen.List(
							jen.Id("ret"),
							jen.Qual("errors", "New").Call(
								jen.Lit(fmt.Sprintf("unable to coerce value for field '%s'", name)),
							),
						)),
					)
					g.Return(jen.List(jen.Id("ret"), jen.Id("err")))
				} else {
					g.Return(jen.List(jen.Id("val"), jen.Id("err")))
				}
			})
	}
	return code
}

func mkFieldResolversName(prefix string) string {
	return prefix + "FieldResolvers"
}
