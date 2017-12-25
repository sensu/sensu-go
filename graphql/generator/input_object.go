package generator

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genInputObject(node *ast.InputObjectDefinition) jen.Code {
	code := newGroup()
	name := node.GetName().Value

	// Type description
	desc := getDescription(node)
	comment := genTypeComment(name, desc)

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
	code.Comment(comment)
	code.Func().Id(name).Params().Qual(defsPkg, "InputObjectConfig").Block(
		jen.Return(jen.Qual(defsPkg, "InputObjectConfig").Values(jen.Dict{
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(desc),
			jen.Id("Fields"):      genInputObjectFields(node.Fields),
		})),
	)

	return code
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
	//    graphql.InputObjectFieldConfigMap{
	//      "name":        &graphql.InputObjectFieldConfig{ ... },
	//      "newInterval": &graphql.InputObjectFieldConfig{ ... },
	//    }
	//

	return jen.Qual(defsPkg, "InputObjectConfigFieldMap").Values(
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

	return jen.Op("&").Qual(defsPkg, "InputObjectFieldConfig").Values(jen.Dict{
		jen.Id("Type"):         genInputTypeReference(field.Type),
		jen.Id("Description"):  genDescription(field),
		jen.Id("DefaultValue"): genValue(field.DefaultValue),
	})
}
