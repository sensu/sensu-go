package generator

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genInterface(node *ast.InterfaceDefinition) jen.Code {
	code := newGroup()
	name := node.GetName().Value
	resolverName := fmt.Sprintf("%sResolver", name)

	//
	// Generate resolver interface
	//
	// ... comment: Describe resolver interface and usage
	// ... method:  ResolveType
	//

	code.Commentf(`//
// %s represents a collection of methods whose products represent the input and
// response values of a interface type.
//
//  == Example generated interface
//
//  // PetResolver ...
//  type PetResolver interface {
//    // ResolveType should return name of type given a value
//    ResolveType(interface{}, graphql.ResolveTypeParams) string
//  }
//
//  // Example implementation ...
//
//  // MyPetResolver implements DateResolver interface
//  type MyPetResolver struct {
//    logger    logrus.LogEntry
//  }
//
//  // ResolveType should return name of type given a value
//  func (r *MyPetResolver) ResolveType(val interface {}, _ graphql.ResolveTypeParams) string {
//    // ... implementation details ...
//    switch pet := val.(type) {
//    when *Dog:
//      return "Dog" // Handled by type identified by 'Dog'
//    when *Cat:
//      return "Cat" // Handled by type identified by 'Cat'
//    }
//    panic("Unimplemented")
//  }`,
		resolverName,
	)
	// Generate resolver interface.
	code.Type().Id(resolverName).Interface(
		jen.Comment("ResolveType should return name of type given a value"),
		jen.Id("ResolveType").Params(jen.Qual(graphqlGoPkg, "ResolveTypeParams")).Op("*").String(),
	)

	//
	// Generate type definition
	//
	// ... comment: Include description in comment
	// ... panic callbacks panic if not configured
	//

	// Interface description
	typeDesc := fetchDescription(node)

	// To appease the linter ensure that the the description of the scalar begins
	// with the name of the resulting method.
	desc := typeDesc
	if hasPrefix := strings.HasPrefix(typeDesc, name); !hasPrefix {
		desc = name + " " + desc
	}

	//
	// Generates thunk that returns new instance of interface config
	//
	//  == Example input SDL
	//
	//    "Pets are the bestest family members"
	//    interface Pet {
	//      "name of this fine beast."
	//      name: String!
	//    }
	//
	//  == Example output
	//
	//    // Pets are the bestest family members
	//    func Pet() graphql.InterfaceConfig { // implements TypeThunk
	//      return graphql.InterfaceConfig{
	//        Name:        "Pet",
	//        Description: "Pets are the bestest family members",
	//        Fields:      // ...
	//        ResolveType: func (_ ResolveTypeParams) string {
	//          panic("Unimplemented; see PetResolver.")
	//        },
	//      }
	//    }
	//
	code.Comment(desc)
	code.Func().Id(name).Params().Qual(graphqlGoPkg, "InterfaceConfig").Block(
		jen.Return(jen.Qual(graphqlGoPkg, "InterfaceConfig").Values(jen.Dict{
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(typeDesc),
			jen.Id("Fields"):      genFields(node.Fields),
			jen.Id("ResolveType"): jen.Func().
				Params(jen.Id("_").Qual(graphqlGoPkg, "ResolveTypeParams")).
				Op("*").Qual(graphqlGoPkg, "Object").
				Block(
					jen.Comment(missingResolverNote),
					jen.Panic(jen.Lit("Unimplemented; see "+resolverName+".")),
				),
		})),
	)

	return code
}
