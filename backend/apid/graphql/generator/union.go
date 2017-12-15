package generator

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genUnion(f *jen.File, node *ast.ScalarDefinition) error {
	name := node.GetName().Value
	resolverName := fmt.Sprintf("%sResolver", name)

	//
	// Generate resolver interface
	//
	// ... comment: Describe resolver interface and usage
	// ... method:  ResolveType
	//

	// 	f.Commentf(`//
	// // %s represents a collection of methods whose products represent the input and
	// // response values of a scalar type.
	// //
	// //  == Example generated interface
	// //
	// //  // DateResolver ...
	// //  type DateResolver interface {
	// //    // ResolveType ... TODO
	// //    ResolveType(graphql.ResolveTypeParams) *graphql.Object
	// //  }
	// //
	// //  // Example implementation ...
	// //
	// //  // MyDateResolver implements DateResolver interface
	// //  type MyDateResolver struct {
	// //    defaultTZ *time.Location
	// //    logger    logrus.LogEntry
	// //  }
	// //
	// //  // ResolveType ... TODO
	// //  func (r *MyDateResolver) ResolveType(p graphql.ResolveTypeParams) *graphql.Object {
	// //    // ... implementation details ...
	// //  }`,
	// 		resolverName,
	// 	)
	// 	// Generate resolver interface.
	// 	f.Type().Id(resolverName).Interface(
	// 		// Serialize method.
	// 		jen.Comment("Serialize an internal value to include in a response."),
	// 		jen.Id("Serialize").Params(jen.Id("interface{}")).Interface(),
	// 	)

	//
	// Generate type definition
	//
	// ... comment: Include description in comment
	// ... panic callbacks panic if not configured
	//

	// Union description
	typeDesc := fetchDescription(node)

	// To appease the linter ensure that the the description of the scalar begins
	// with the name of the resulting method.
	desc := typeDesc
	if hasPrefix := strings.HasPrefix(typeDesc, name); !hasPrefix {
		desc = name + " " + desc
	}

	// Ex.
	//   // NameOfMyUnion [the description given in SDL document]
	//   func NameOfMyUnion() *graphql.Scalar { ... } // implements TypeThunk
	f.Comment(desc)
	f.Func().Id(name).Params().Op("*").Qual(graphqlPkg, "Union").Block(
		jen.Return(jen.Qual(graphqlPkg, "UnionConfig").Values(jen.Dict{
			// Name & description
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(typeDesc),
			jen.Id("Types"):       jen.Lit(typeDesc),

			// Resolver funcs
			jen.Id("Serialize"): jen.Func().Params(jen.Id("_").Interface()).Block(
				jen.Comment(missingResolverNote),
				jen.Panic(jen.Lit("Unimplemented; see "+resolverName+".")),
			),
		})),
	)

	return nil
}
