package generator

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genUnion(f *jen.File, node *ast.UnionDefinition) error {
	name := node.GetName().Value
	resolverName := fmt.Sprintf("%sResolver", name)

	//
	// Generate resolver interface
	//
	// ... comment: Describe resolver interface and usage
	// ... method:  ResolveType
	//

	f.Commentf(`//
// %s represents a collection of methods whose products represent the 
// response values of a union type.
//
//  == Example generated interface
//
//  // FeedResolver ...
//  type FeedResolver interface {
//    // ResolveType should return name of type given a value
//    ResolveType(graphql.ResolveTypeParams) string
//  }
//
//  // Example implementation ...
//
//  // MyFeedResolver implements FeedResolver interface
//  type MyFeedResolver struct {
//    logger    logrus.LogEntry
//  }
//
//  // ResolveType ... TODO
//  func (r *MyFeedResolver) ResolveType(p graphql.ResolveTypeParams) *graphql.Object {
//    // ... implementation details ...
//  }`,
		resolverName,
	)
	// Generate resolver interface.
	f.Type().Id(resolverName).Interface(
		// ResolveType method.
		jen.Comment("ResolveType should return name of type given a value."),
		jen.Id("ResolveType").Params(jen.Qual(graphqlPkg, "ResolveTypeParams")).String(),
	)

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
	f.Func().Id(name).Params().Qual(graphqlPkg, "UnionConfig").Block(
		jen.Return(jen.Qual(graphqlPkg, "UnionConfig").Values(jen.Dict{
			// Name & description
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(typeDesc),
			jen.Id("Types"): jen.Index().Op("*").Qual(graphqlPkg, "Object").Values(
				jen.ValuesFunc(func(g *jen.Group) {
					for _, t := range node.Types {
						g.Line().Op("&").Qual(graphqlPkg, "Object").Values(jen.Dict{
							jen.Id("PrivateName"): jen.Lit(t.Name.Value),
						})
					}
				}),
			),

			// Resolver funcs
			jen.Id("ResolveType"): jen.Func().Params(jen.Id("_").Qual(graphqlPkg, "ResolveTypeParams")).String().Block(
				jen.Comment(missingResolverNote),
				jen.Panic(jen.Lit("Unimplemented; see "+resolverName+".")),
			),
		})),
	)

	return nil
}
