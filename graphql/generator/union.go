package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/graphql-go/graphql/language/ast"
)

//
// Generates description for union type
//
// == Example input SDL
//
//   """
//   Feed includes all stuff and things.
//   """
//   union Feed = Story | Article | Advert
//
// == Example implementation
//
//   // FeedType - Feed includes all the stuff and things.
//   var FeedType = graphql.NewType("Feed", graphql.UnionKind)
//
//   // RegisterFeed registers Feed union type with given service.
//   func RegisterFeed(svc graphql.Service, impl graphql.UnionTypeResolver) {
//     svc.RegisterUnion(_UnionTypeFeedDesc, impl)
//   }
//
//   // define configuration thunk
//   func _UnionTypeFeedConfigFn() graphql.UnionConfig {
//     return graphql.UnionConfig{
//       Name:        "Feed",
//       Description: "Feed includes all stuff and things.",
//       Types:       // ...
//       ResolveType: func (_ ResolveTypeParams) string {
//         panic("Unimplemented; see UnionTypeResolver.")
//       },
//     }
//   }
//
//   // describe feed's configuration; kept private to avoid unintentional
//   // tampering at runtime.
//   var _UnionTypeFeedDesc = graphql.UnionDesc{
//     Config: _UnionTypeFeedConfigFn,
//   }
//
func genUnion(node *ast.UnionDefinition) jen.Code {
	code := newGroup()
	name := node.GetName().Value

	// Type description
	desc := getDescription(node)

	// Ids
	publicRefName := name + "Type"
	publicRefComment := genTypeComment(publicRefName, desc)
	privateConfigName := mkPrivateID(node, "Desc")
	privateConfigThunkName := mkPrivateID(node, "ConfigFn")

	//
	// Generate public reference to type
	//
	// == Example output
	//
	//   // FeedType - Feed includes all stuff and things.
	//   var FeedType = graphql.NewType("Feed", graphql.UnionKind)
	//
	code.Comment(publicRefComment)
	code.
		Var().Id(publicRefName).Op("=").
		Qual(servicePkg, "NewType").
		Call(jen.Lit(name), jen.Qual(servicePkg, "UnionKind"))

	//
	// Generate public func to register type with service
	//
	// == Example output
	//
	//   // RegisterFeed registers Feed union type with given service.
	//   func RegisterFeed(svc graphql.Service, impl graphql.UnionTypeResolver) {
	//     svc.RegisterUnion(_UnionTypeFeedDesc, impl)
	//   }
	//

	code.Add(
		genRegisterFn(node, jen.Qual(servicePkg, "UnionTypeResolver")),
	)

	//
	// Generates type config thunk
	//
	// == Example output
	//
	//   // define configuration thunk
	//   func _UnionTypeFeedConfigFn() graphql.UnionConfig {
	//     return graphql.UnionConfig{
	//       Name:        "Feed",
	//       Description: "Feed includes all stuff and things.",
	//       Types:       // ...
	//       ResolveType: func (_ ResolveTypeParams) string {
	//         panic("Unimplemented; see UnionTypeResolver.")
	//       },
	//     }
	//   }
	//

	code.
		Func().Id(privateConfigThunkName).
		Params().Qual(defsPkg, "UnionConfig").
		Block(
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
						jen.Panic(jen.Lit("Unimplemented; see UnionTypeResolver.")),
					),
			})),
		)

	//
	// Generate type description
	//
	// == Example output
	//
	//   // describe feed's configuration; kept private to avoid unintentional
	//   // tampering at runtime.
	//   var _UnionTypeFeedDesc = graphql.UnionDesc{
	//     Config: _UnionTypeFeedConfigFn,
	//   }
	//
	code.Commentf(
		`describe %s's configuration; kept private to avoid unintentional tampering of configuration at runtime.`,
		name,
	)
	code.
		Var().Id(privateConfigName).Op("=").
		Qual(servicePkg, "UnionDesc").
		Values(jen.Dict{
			jen.Id("Config"): jen.Id(privateConfigThunkName),
		})

	return code
}
