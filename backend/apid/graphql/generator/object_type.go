package generator

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genObjectType(f *jen.File, node *ast.ObjectDefinition) error {
	name := node.GetName().Value
	resolverName := fmt.Sprintf("%sResolver", name)

	//
	// Generate resolver interface
	//
	// ... comment: Describe resolver interface and usage
	// ... method:  [one method for each field]
	//

	f.Commentf(`//
// %s represents a collection of methods whose products represent the
// response values of an object type.
//
//  == Example SDL
//
//    """
//    Dog's are not hooman.
//    """
//    type Dog implements Pet {
//      "name of this fine beast."
//      name:  String!
//
//      "breed of this silly animal; probably shibe."
//      breed: [Breed]
//    }
//
//  == Example generated interface
//
//  // DateResolver ...
//  type DogResolver interface {
//    // Name implements response to request for name field.
//    Name(context.Context, interface{}, graphql.Params) interface{}
//    // Breed implements response to request for breed field.
//    Breed(context.Context, interface{}, graphql.Params) interface{}
//    // IsTypeOf is used to determine if a given value is associated with the Dog type
//    IsTypeOf(context.Context, graphql.IsTypeOfParams) bool
//  }
//
//  == Example implementation ...
//
//  // MyDogResolver implements DateResolver interface
//  type MyDogResolver struct {
//    logger logrus.LogEntry
//    store interface{
//      store.BreedStore
//      store.DogStore
//    }
//  }
//
//  // Name implements response to request for name field.
//  func (r *MyDogResolver) Name(ctx context.Context, r interface{}, p graphql.Params) interface{} {
//    // ... implementation details ...
//    dog := r.(DogGetter)
//    return dog.GetName()
//  }
//
//  // Breed implements response to request for breed field.
//  func (r *MyDogResolver) Name(ctx context.Context, r interface{}, p graphql.Params) interface{} {
//    // ... implementation details ...
//    dog := r.(DogGetter)
//    breed := r.store.GetBreed(dog.GetBreedName())
//    return breed
//  }`,
		resolverName,
	)
	// Generate resolver interface.
	f.Type().Id(resolverName).InterfaceFunc(func(g *jen.Group) {
		for _, field := range node.Fields {
			// Define method for each field in object type
			name := field.Name.Value
			titleizedName := strings.Title(field.Name.Value)

			// intended interface is (context.Context, record interface{}, params graphql.Params).
			// while this differs from graphql packages I feel it is more conventional.
			g.Commentf("%s implements response to request for '%s' field.", name, titleizedName)
			g.Id(titleizedName).Params(
				jen.Qual("context", "Context"),
				jen.Id("interface{}"),
				jen.Qual(graphqlPkg, "Params"),
			).Interface()
		}

		// Satisfy IsTypeOf() callback
		g.Commentf("IsTypeOf is used to determine if a given value is associated with the %s type", name)
		g.Id("IsTypeOf").Params( // IsTypeOf(context.Context, graphql.IsTypeOfParams) bool
			jen.Qual("context", "Context"),
			jen.Qual(graphqlPkg, "IsTypeOfParams"),
		).Interface()
	})

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
	ints := jen.Index().Op("*").Qual(graphqlPkg, "Interface").Values(
		jen.ValuesFunc(func(g *jen.Group) {
			for _, in := range node.Interfaces {
				g.Line().Op("&").Qual(graphqlPkg, "Interface").Values(jen.Dict{
					jen.Id("PrivateName"): jen.Lit(in.Name.Value),
				})
			}
		}),
	)

	//
	// Generates thunk that returns new instance of object type
	//
	//  == Example input SDL
	//
	//    """
	//    Dogs are not hooman.
	//    """
	//    type Dog implements Pet {
	//      "name of this fine beast."
	//      name:  String!
	//
	//      "breed of this silly animal; probably shibe."
	//      breed: [Breed]
	//    }
	//
	//   == Example output
	//
	//   // Dogs are not hooman
	//   func Dog() *graphql.Object { // implements TypeThunk
	//     return graphql.NewObject(graphql.ObjectConfig{
	//       Name:        "Dog",
	//       Description: "are not hooman",
	//       Interfaces:  // ...
	//       Fields:      // ...
	//       IsKindOf:    // ...
	//     })
	//   }
	//
	f.Comment(desc)
	f.Func().Id(name).Params().Qual(graphqlPkg, "ObjectConfig").Block(
		jen.Return(jen.Qual(graphqlPkg, "ObjectConfig").Values(jen.Dict{
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(typeDesc),
			jen.Id("Interfaces"):  ints,
			jen.Id("Fields"):      genFields(node.Fields),
			jen.Id("IsKindOf"): jen.Func().Params(jen.Id("_").Qual(astPkg, "Value")).Block(
				jen.Comment(missingResolverNote),
				jen.Panic(jen.Lit("Unimplemented; see "+resolverName+".")),
			),
		})),
	)
	return nil
}
