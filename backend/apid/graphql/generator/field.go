package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

// genFields generates fields config for given AST
func genFields(fs []*ast.FieldDefinition) *jen.Statement {
	//
	// Generate config for fields
	//
	//  == Example input SDL
	//
	//    type Dog {
	//      name(style: NameComponentsStyle = SHORT): String!
	//      givenName: String @deprecated(reason: "No longer supported; please use name field.")
	//    }
	//
	//  == Example output
	//
	//    graphql.Fields{
	//      "name":      graphql.Field{ ... },
	//      "givenName": graphql.Field{ ... },
	//    }
	//
	return jen.Qual(graphqlPkg, "Fields").Values(jen.DictFunc(func(d jen.Dict) {
		for _, f := range fs {
			d[jen.Id(f.Name.Value)] = genField(f)
		}
	}))
}

// genField generates field config for given AST
func genField(field *ast.FieldDefinition) *jen.Statement {
	//
	// Generate config for field
	//
	//  == Example input SDL
	//
	//    interface Pet {
	//      "name of the pet"
	//      name(style: NameComponentsStyle = SHORT): String!
	//      """
	//      givenName of the pet ★
	//      """
	//      givenName: String @deprecated(reason: "No longer supported; please use name field.")
	//    }
	//
	//  == Example output
	//
	//    &graphql.Field{
	//      Name:              "name",
	//      Type:              graphql.NonNull(graphql.String),
	//      Description:       "name of the pet",
	//      DeprecationReason: "",
	//      Args:              FieldConfigArgument{ ... },
	//    }
	//
	//    &graphql.Field{
	//      Name:              "givenName",
	//      Type:              graphql.String,
	//      Description:       "givenName of the pet",
	//      DeprecationReason: "No longer supported; please use name field.",
	//      Args:              FieldConfigArgument{ ... },
	//    }
	//

	// TODO: Match concrete types (Int, String, etc.)
	// TODO: Match lists, nonnull, etc.
	ttype := field.Type.String()

	depReason := fetchDeprecationReason(field.Directives)
	return jen.Qual(graphqlPkg, "Field").Values(jen.Dict{
		jen.Id("Args"):              genArguments(field),
		jen.Id("DeprecationReason"): jen.Lit(depReason),
		jen.Id("Description"):       jen.Lit(fetchDescription(field)),
		jen.Id("Name"):              jen.Lit(field.Name.Value),
		jen.Id("Type"):              jen.Qual(utilPkg, "OutputType").Call(jen.Lit(ttype)),
	})
}

// genField generates field config for given AST
func genArguments(f *ast.FieldDefinition) *jen.Statement {
	return jen.Qual(graphqlPkg, "FieldConfigArgument").Values(
		jen.DictFunc(func(d jen.Dict) {
			for _, arg := range f.Arguments {
				// TODO: Match concrete types (Int, String, etc.)
				// TODO: Match lists, nonnull, etc.
				ttype := arg.Type.String()
				def := jen.Op("*").Qual(graphqlPkg, "ArgumentConfig").Values(jen.Dict{
					jen.Id("Type"): jen.Qual(utilPkg, "InputType").Call(jen.Lit(ttype)),
				})

				d[jen.Id(arg.Name.Value)] = def
			}
		}),
	)
}
