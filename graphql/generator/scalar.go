package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/graphql-go/graphql/language/ast"
)

const scalarResolverName = "ScalarResolver"

//
// Generates description for scalar type
//
// == Example input SDL
//
//   """
//   Timestamps are great.
//   """
//   scalar Timestamp
//
// == Example output
//
//   // Timestamps are great
//   var TimestampType = graphql.NewType("Timestamp", graphql.ScalarKind)
//
//   // RegisterTimestamp registers Timestamp scalar type with given service.
//   func RegisterTimestamp(svc graphql.Service, impl graphql.ScalarResolver) {
//     svc.RegisterScalar(_ScalarTypeTimestampDesc, impl)
//   }
//
//   // describe timestamp's configuration; keep private to avoid
//   // unintentional tampering at runtime.
//   var _ScalarTypeTimestampDesc = graphql.ScalarDesc{
//     Config: func() definition.ScalarConfig {
//       return definition.ScalarConfig{
//         Name:         "Timestamp",
//         Description:  "Timestamps are great.",
//         Serialize:    // ...
//         ParseValue:   // ...
//         ParseLiteral: // ...
//       }
//     }
//   }
//
func genScalar(node *ast.ScalarDefinition) jen.Code {
	code := newGroup()
	name := node.GetName().Value

	// Ids
	privateConfigName := mkPrivateID(node, "Desc")
	//privateConfigThunkName := mkPrivateID(node, "ConfigFn")

	// Scalar description
	desc := getDescription(node)
	comment := genTypeComment(name, desc)

	//
	// Generate type config thunk
	//
	// == Example output
	//
	//   func() definition.ScalarConfig {
	//     return definition.ScalarConfig{
	//       Name:         "Timestamp",
	//       Description:  "Timestamps are great.",
	//       Serialize:    // ... closure that panics ...
	//       ParseValue:   // ... closure that panics ...
	//       ParseLiteral: // ... closure that panics ...
	//     }
	//   }
	//
	thunk := jen.Func().Params().Qual(defsPkg, "ScalarConfig").Block(
		jen.Return(jen.Qual(defsPkg, "ScalarConfig").Values(jen.Dict{
			// Name & description
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(desc),

			// Resolver funcs
			jen.Id("Serialize"): jen.Func().Params(jen.Id("_").Interface()).Interface().Block(
				jen.Comment(missingResolverNote),
				jen.Panic(jen.Lit("Unimplemented; see "+scalarResolverName+".")),
			),
			jen.Id("ParseValue"): jen.Func().Params(jen.Id("_").Interface()).Interface().Block(
				jen.Comment(missingResolverNote),
				jen.Panic(jen.Lit("Unimplemented; see "+scalarResolverName+".")),
			),
			jen.Id("ParseLiteral"): jen.Func().Params(jen.Id("_").Qual(astPkg, "Value")).Interface().Block(
				jen.Comment(missingResolverNote),
				jen.Panic(jen.Lit("Unimplemented; see "+scalarResolverName+".")),
			),
		})),
	)

	//
	// Generate public reference to type
	//
	// == Example output
	//
	//   // Timestamps are great
	//   var TimestampType = graphql.NewType("Timestamp", graphql.ScalarKind)
	//
	code.Commentf("%s ... %s", name+"Type", comment)
	code.
		Var().Id(name+"Type").Op("=").
		Qual(servicePkg, "NewType").
		Call(jen.Lit(name), jen.Qual(servicePkg, "ScalarKind"))

	//
	// Generate public func to register type with service
	//
	// == Example output
	//
	//   // RegisterTimestamp registers Timestamp scalar type with given service.
	//   func RegisterTimestamp(svc graphql.Service, impl graphql.ScalarResolver) {
	//     svc.RegisterScalar(_ScalarType_Timestamp, impl)
	//   }
	//

	code.Add(
		genRegisterFn(node, jen.Qual(servicePkg, "ScalarResolver")),
	)

	//
	// Generate type description
	//
	// == Example output
	//
	//   // describe timestamp's configuration; kept private to avoid
	//   // unintentional tampering at runtime.
	//   var _ScalarType_Timestamp = graphql.ScalarDesc{
	//     Config: func() definition.ScalarConfig { ... }
	//   }
	//
	code.Commentf(
		`describe %s's configuration; kept private to avoid unintentional tampering of configuration at runtime.`,
		name,
	)
	code.
		Var().Id(privateConfigName).Op("=").
		Qual(servicePkg, "ScalarDesc").
		Values(jen.Dict{jen.Id("Config"): thunk})

	return code
}
