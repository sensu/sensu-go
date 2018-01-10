package generator

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

//
// Generates description for schema
//
// == Example input SDL
//
//   schema {
//     query:    QueryType
//     mutation: MutationType
//     subscription: SubscriptionType
//   }
//
// == Example output
//
//   // Schema supplies the root types of each type of operation, query,
//   // mutation (optional), and subscription (optional).
//   var Schema = graphql.NewType("Schema", graphql.SchemaKind)
//
//   // RegisterSchema registers schema description with given service.
//   func RegisterSchema(svc graphql.Service) {
//     svc.RegisterSchema(_SchemaDesc)
//   }
//
//   // define configuration thunk
//   func _SchemaConfigFn() graphql.SchemaConfig {
//     return graphql.SchemaConfig{
//       Query:        &graphql.Object{PrivateName: "QueryType"},
//       Mutation:     &graphql.Object{PrivateName: "MutationType"},
//     }
//   }
//
//   // describe schema's configuration; keep private to avoid
//   // unintentional tampering at runtime.
//   func _SchemaDesc = graphql.SchemaDesc{
//     Config: _SchemaConfigFn,
//   }
//
func genSchema(node *ast.SchemaDefinition) jen.Code {
	code := newGroup()

	// Ids
	registerFnName := "RegisterSchema"
	publicRefName := "Schema"
	privateConfigName := "_SchemaDesc"
	privateConfigThunkName := "_SchemaConfigFn"

	//
	// Generate record that will be used to refer to the schema
	//
	// == Example output
	//
	//   // Schema supplies the root types of each type of operation, query,
	//   // mutation (optional), and subscription (optional).
	//   var Schema = graphql.NewType("Schema", graphql.SchemaKind)
	//
	code.Comment(`// Schema supplies the root types of each type of operation, query,
// mutation (optional), and subscription (optional).`)
	code.
		Var().Id(publicRefName).Op("=").
		Qual(servicePkg, "NewType").
		Call(jen.Lit("Schema"), jen.Qual(servicePkg, "SchemaKind"))

	//
	// Generate public func to register type with service
	//
	// == Example output
	//
	//   // RegisterSchema registers schema description with given service.
	//   func RegisterSchema(svc graphql.Service) {
	//     svc.RegisterSchema(_SchemaDesc)
	//   }
	//
	code.Comment(
		"RegisterSchema registers schema description with given service.",
	)
	code.
		Func().Id(registerFnName).
		Params(jen.Id("svc").Qual(servicePkg, "Service")).
		Block(
			jen.Id("svc.RegisterSchema").Call(
				jen.Id(privateConfigName),
			),
		)

	//
	// Generate type config thunk
	//
	// == Example output
	//
	//   // define configuration thunk
	//   func _SchemaConfigFn() graphql.SchemaConfig {
	//     return graphql.SchemaConfig{
	//       Query:        &graphql.Object{PrivateName: "QueryType"},
	//       Mutation:     &graphql.Object{PrivateName: "MutationType"},
	//     }
	//   }
	//
	code.
		Func().Id(privateConfigThunkName).
		Params().Qual(defsPkg, "SchemaConfig").
		Block(
			jen.Return(jen.Qual(defsPkg, "SchemaConfig").Values(
				jen.DictFunc(func(d jen.Dict) {
					for _, op := range node.OperationTypes {
						opName := strings.Title(op.Operation)
						d[jen.Id(opName)] = genMockObjectReference(op.Type)
					}
				}),
			)),
		)

	//
	// Generate type description
	//
	// == Example output
	//
	//   // describe schema's configuration; keep private to avoid
	//   // unintentional tampering at runtime.
	//   func _SchemaDesc = graphql.SchemaDesc{
	//     Config: _SchemaConfigFn,
	//   }
	//
	code.Comment(
		`describe schema's configuration; kept private to avoid unintentional tampering of configuration at runtime.`,
	)
	code.
		Var().Id(privateConfigName).Op("=").
		Qual(servicePkg, "SchemaDesc").
		Values(jen.Dict{
			jen.Id("Config"): jen.Id(privateConfigThunkName),
		})

	return code
}
