package generator

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genInputObject(f *jen.File, node *ast.InputObjectDefinition) error {
	name := node.GetName().Value
	typeDesc := fetchDescription(node)

	// To appease the linter ensure that the the description of the object begins
	// with the name of the resulting method.
	desc := typeDesc
	if hasPrefix := strings.HasPrefix(typeDesc, name); !hasPrefix {
		desc = name + " " + desc
	}

	//
	// Generate input object config
	//
	//  == Example input SDL
	//
	//    """
	//    ConfigureIntervalInput ...
	//    """
	//    input ConfigureIntervalInput {
	//      "name of the check"
	//      name: String!
	//
	//      "new interval in seconds"
	//      newInterval: Int!
	//    }
	//
	//  == Example output
	//
	//   func ConfigureIntervalInput() graphql.InputObjectConfig {
	//     return graphql.InputObjectConfig{
	//       Name: "ConfigureIntervalInput",
	//       Description: "self descriptive",
	//       Fields: InputObjectFieldMap{ ... },
	//    }
	//  }
	//
	f.Comment(desc)
	f.Func().Id(name).Params().Qual(graphqlPkg, "InputObjectConfig").Block(
		jen.Return(jen.Qual(graphqlPkg, "InputObjectConfig").Values(jen.Dict{
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(typeDesc),
			jen.Id("Fields"):      genInputObjectFields(node.Fields),
		})),
	)

	return nil
}

func genInputObjectFields(fields []*ast.InputValueDefinition) jen.Code {
	//
	// Generate input object fields config
	//
	//  == Example input SDL
	//
	//    """
	//    ConfigureIntervalInput ...
	//    """
	//    input ConfigureIntervalInput {
	//      "name of the check"
	//      name: String!
	//
	//      "new interval in seconds"
	//      newInterval: Int!
	//    }
	//
	//  == Example output
	//
	//    graphql.InputObjectFieldMap{
	//      "name":        &graphql.InputObjectFieldConfig{ ... },
	//      "newInterval": &graphql.InputObjectFieldConfig{ ... },
	//    }
	//

	return jen.Qual(graphqlPkg, "InputObjectFieldMap").Values(
		jen.DictFunc(func(d jen.Dict) {
			for _, field := range fields {
				d[jen.Lit(field.Name.Value)] = genInputObjectField(field)
			}
		}),
	)
}

func genInputObjectField(field *ast.InputValueDefinition) jen.Code {
	//
	// Generate input object fields config
	//
	//  == Example input SDL
	//
	//    """
	//    ConfigureIntervalInput ...
	//    """
	//    input ConfigureIntervalInput {
	//      "name of the check"
	//      name: String!
	//
	//      "new interval in seconds"
	//      newInterval: Int = 60!
	//    }
	//
	//  == Example output
	//
	//    &graphql.InputObjectFieldConfig{
	//      Name: "newInterval",
	//      Type: graphql.NonNull(graphql.Int),
	//      DefaultValue: 60,
	//    },
	//

	desc := fetchDescription(field)
	return jen.Op("&").Qual(graphqlPkg, "InputObjectFieldConfig").Values(jen.Dict{
		jen.Lit("Name"):        jen.Lit(field.Name.Value),
		jen.Lit("Type"):        genInputTypeReference(field.Type),
		jen.Lit("Description"): jen.Lit(desc),
	})
}
