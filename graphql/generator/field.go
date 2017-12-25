package generator

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

// Titleizes given name to match
func toFieldName(name string) string {
	name = strings.Title(name)

	// NOTE: golint prefers method names use "ID" instead of "Id".
	if name == "Id" {
		name = "ID"
	}
	return name
}

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
	return jen.Qual(defsPkg, "Fields").Values(jen.DictFunc(func(d jen.Dict) {
		for _, f := range fs {
			d[jen.Lit(f.Name.Value)] = genField(f)
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
	return jen.Op("&").Qual(defsPkg, "Field").Values(jen.Dict{
		jen.Id("Args"):              genArguments(field.Arguments),
		jen.Id("DeprecationReason"): genDeprecationReason(field.Directives),
		jen.Id("Description"):       genDescription(field),
		jen.Id("Name"):              jen.Lit(field.Name.Value),
		jen.Id("Type"):              genOutputTypeReference(field.Type),
	})
}

// genArguments generates argument field config for given AST
func genArguments(args []*ast.InputValueDefinition) *jen.Statement {
	//
	// Generate config for arguments
	//
	//  == Example input SDL
	//
	//    type Dog {
	//      name(
	//       "style is stylish"
	//       style: NameComponentsStyle = SHORT,
	//      ): String!
	//    }
	//
	//  == Example output
	//
	//    FieldConfigArgument{
	//      "style": &ArgumentConfig{ ... }
	//    },
	//
	return jen.Qual(defsPkg, "FieldConfigArgument").Values(
		jen.DictFunc(func(d jen.Dict) {
			for _, arg := range args {
				d[jen.Lit(arg.Name.Value)] = genArgument(arg)
			}
		}),
	)
}

// genArgument generates argument config for given AST
func genArgument(arg *ast.InputValueDefinition) *jen.Statement {
	//
	// Generate config for argument
	//
	//  == Example input SDL
	//
	//    type Dog {
	//      name(
	//       "style is stylish"
	//       style: NameComponentsStyle = SHORT,
	//      ): String!
	//    }
	//
	//  == Example output
	//
	//    &ArgumentConfig{
	//      Type: graphql.NonNull(graphql.String),
	//      DefaultValue: "SHORT", // TODO: ???
	//      Description: "style is stylish",
	//    }
	//
	return jen.Op("&").Qual(defsPkg, "ArgumentConfig").Values(jen.Dict{
		jen.Id("DefaultValue"): genValue(arg.DefaultValue),
		jen.Id("Description"):  genDescription(arg),
		jen.Id("Type"):         genInputTypeReference(arg.Type),
	})
}

func genValue(v ast.Value) jen.Code {
	switch val := v.(type) {
	case nil:
		return jen.Null()
	case *ast.ListValue:
		return jen.Index().Interface().ValuesFunc(func(g *jen.Group) {
			for _, lval := range val.Values {
				g.Add(genValue(lval))
			}
		})
	case *ast.ObjectValue:
		return jen.Map(jen.String()).Interface().Values(
			jen.DictFunc(func(d jen.Dict) {
				for _, f := range val.Fields {
					d[jen.Lit(f.Name.Value)] = genValue(f.Value)
				}
			}),
		)
	}
	return jen.Lit(v.GetValue())
}
