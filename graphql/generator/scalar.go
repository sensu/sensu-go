package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

const scalarResolverName = "ScalarResolver"

//
// Generates thunk that returns new instance of scalar config
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
//   var Timestamp = graphql.NewType("Timestamp", graphql.ScalarKind)
//
//   // RegisterTimestamp registers Timestamp scalar type with given service.
//   func RegisterTimestamp(svc graphql.Service, impl graphql.ScalarResolver) {
//     src.RegisterScalar(_ScalarType_Timestamp, impl)
//   }
//
//   // describe timestamp's configuration; keep private to avoid
//   // unintentional tampering at runtime.
//   var _ScalarType_Timestamp = graphql.ScalarDesc{
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
	privateName := "_ScalarType_" + name

	// Scalar description
	desc := getDescription(node)
	comment := genTypeComment(node)

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
	thunk := jen.Func().Id(name).Params().Qual(defsPkg, "ScalarConfig").Block(
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
	//   var Timestamp = graphql.NewType("Timestamp", graphql.ScalarKind)
	//
	code.Comment(comment)
	code.
		Var().Lit(name).Op("=").
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
	registerFnName := fmt.Sprintf("Register%s", name)
	code.Commentf(
		"Register%s registers %s scalar type with given service.",
		registerFnName,
		name,
	)
	code.
		Func().Id(registerFnName).
		Params(
			jen.Id("svc").Qual(servicePkg, "Service"),
			jen.Id("impl").Qual(servicePkg, "ScalarResolver"),
		).
		Block(
			jen.Id("svc.RegisterScalar").Call(
				jen.Id(privateName),
				jen.Id("impl"),
			),
		)

	//
	// Generate type description
	//
	// == Example output
	//
	//   // describe timestamp's configuration; keep private to avoid
	//   // unintentional tampering at runtime.
	//   var _ScalarType_Timestamp = graphql.ScalarDesc{
	//     Config: func() definition.ScalarConfig { ... }
	//   }
	//
	code.Commentf(
		`describe %s's configuration; keep private to avoid unintentional tampering of configuration at runtime.`,
		registerFnName,
		name,
	)
	code.
		Var().Id(privateName).Op("=").
		Qual(servicePkg, "ScalarDesc").
		Values(jen.Dict{jen.Lit("Config"): thunk})

	return code
}
