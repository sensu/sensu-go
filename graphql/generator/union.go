package generator

import (
	"fmt"

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
		jen.Id("ResolveType").Params(jen.Qual(defsPkg, "ResolveTypeParams")).String(),
	)

	//
	// Generate type definition
	//
	// ... comment: Include description in comment
	// ... panic callbacks panic if not configured
	//

	// Type description
	desc := getDescription(node)
	comment := genTypeComment(name, desc)

	// Ex.
	//   // NameOfMyUnion [the description given in SDL document]
	//   func NameOfMyUnion() *graphql.Scalar { ... } // implements TypeThunk
	code.Comment(comment)
	code.Func().Id(name).Params().Qual(defsPkg, "UnionConfig").Block(
		jen.Return(jen.Qual(defsPkg, "UnionConfig").Values(jen.Dict{
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(desc),
			jen.Id("Types"): jen.Index().Op("*").Qual(defsPkg, "Object").ValuesFunc(
				func(g *jen.Group) {
					for _, t := range node.Types {
						g.Line().Add(genMockObjectReference(t))
					}
				},
			),
			jen.Id("ResolveType"): jen.Func().
				Params(jen.Id("_").Qual(defsPkg, "ResolveTypeParams")).
				Op("*").Qual(defsPkg, "Object").
				Block(
					jen.Comment(missingResolverNote),
					jen.Panic(jen.Lit("Unimplemented; see "+resolverName+".")),
				),
		})),
	)

	return code
}
