package generator

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genSchema(node *ast.SchemaDefinition) jen.Code {
	code := newGroup()

	//
	// Generates thunk that returns new instance of schema config
	//
	//  == Example input SDL
	//
	//    schema {
	//      query:    QueryType
	//      mutation: MutationType
	//
	//      // TODO: To be implemented.
	//      // subscription: SubscriptionType
	//    }
	//
	//  == Example output
	//
	//    // Schema exposes the root types for each operation.
	//    func Schema() graphql.SchemaConfig {
	//      return graphql.SchemaConfig{
	//        Query:        &graphql.Object{PrivateName: "QueryType"},
	//        Mutation:     &graphql.Object{PrivateName: "MutationType"},
	//      }
	//    }
	//
	code.Comment("Schema exposes the root types for each operation.")
	code.Func().Id("Schema").Params().Qual(graphqlGoPkg, "SchemaConfig").Block(
		jen.Return(jen.Qual(graphqlGoPkg, "SchemaConfig").Values(
			jen.DictFunc(func(d jen.Dict) {
				for _, op := range node.OperationTypes {
					opName := strings.Title(op.Operation)
					d[jen.Id(opName)] = genMockObjectReference(op.Type)
				}
			}),
		)),
	)

	return code
}
