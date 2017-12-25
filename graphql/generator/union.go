package generator

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genUnion(node *ast.UnionDefinition) jen.Code {
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
	code.Type().Id(resolverName).Interface(
		// ResolveType method.
		jen.Comment("ResolveType should return name of type given a value."),
		jen.Id("ResolveType").Params(jen.Qual(graphqlGoPkg, "ResolveTypeParams")).String(),
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
	code.Comment(desc)
	code.Func().Id(name).Params().Qual(graphqlGoPkg, "UnionConfig").Block(
		jen.Return(jen.Qual(graphqlGoPkg, "UnionConfig").Values(jen.Dict{
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(typeDesc),
			jen.Id("Types"): jen.Index().Op("*").Qual(graphqlGoPkg, "Object").ValuesFunc(
				func(g *jen.Group) {
					for _, t := range node.Types {
						g.Line().Add(genMockObjectReference(t))
					}
				},
			),
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
