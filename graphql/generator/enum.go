package generator

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genEnum(node *ast.EnumDefinition) jen.Code {
	code := newGroup()
	name := node.GetName().Value

	//
	// Generate type definition
	//
	// ... comment: Include description in comment
	// ... func:    returns enum configuration
	//

	// Enum description
	typeDesc := fetchDescription(node)

	// To appease the linter ensure that the the description of the enum begins
	// with the name of the resulting method.
	desc := typeDesc
	if hasPrefix := strings.HasPrefix(typeDesc, name); !hasPrefix {
		desc = name + " " + desc
	}

	//
	// Generate config for enum
	//
	//  == Example input SDL
	//
	//    """
	//    Locale describes a place with a distinct (and supported) lanugage / dialect.
	//    """
	//    enum Locale {
	//      """
	//      Canada
	//      """
	//      CA
	//      """
	//      Latin
	//      """
	//      LA @deprecated
	//    }
	//
	//  == Example output
	//
	//    // Locale describes a place with a distinct (and supported) lanugage / dialect.
	//    func Locale() graphql.EnumConfig {
	//      return graphql.EnumConfig{
	//        Name:        "Locale",
	//        Description: "Locale describes a place with a distinct (and supported) lanugage / dialect.",
	//        Values: graphql.EnumValueConfigMap{
	//          "CA": &graphql.EnumValueConfig{
	//            Value:             "CA",
	//            Description:       "Canada",
	//            DeprecationReason: "",
	//          },
	//          "LA": &graphql.EnumValueConfig{
	//            Value:             "LA",
	//            Description:       "Latin",
	//            DeprecationReason: "No longer supported,
	//          },
	//        },
	//      }
	//    }
	code.Comment(desc)
	code.Func().Id(name).Params().Qual(graphqlGoPkg, "EnumConfig").Block(
		jen.Return(jen.Qual(graphqlGoPkg, "EnumConfig").Values(jen.Dict{
			// Name & description
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(typeDesc),
			jen.Id("Values"):      genEnumValues(node.Values),
		})),
	)

	return code
}

func genEnumValues(values []*ast.EnumValueDefinition) jen.Code {
	//
	// Generate config for enum
	//
	//  == Example input SDL
	//
	//    """
	//    Locale describes a place with a distinct (and supported) lanugage / dialect.
	//    """
	//    enum Locale {
	//      """
	//      Canada
	//      """
	//      CA
	//      """
	//      Latin
	//      """
	//      LA @deprecated
	//    }
	//
	//  == Example output
	//
	//    graphql.EnumValueConfigMap{
	//      "CA": &graphql.EnumValueConfig{
	//        Value:             "CA",
	//        Description:       "Canada",
	//        DeprecationReason: "",
	//      },
	//      "LA": &graphql.EnumValueConfig{
	//        Value:             "LA",
	//        Description:       "Latin",
	//        DeprecationReason: "No longer supported,
	//      },
	//    }
	//

	return jen.Qual(graphqlGoPkg, "EnumValueConfigMap").Values(
		jen.DictFunc(func(d jen.Dict) {
			for _, v := range values {
				d[jen.Lit(v.Name.Value)] = genEnumValue(v)
			}
		}),
	)
}

func genEnumValue(val *ast.EnumValueDefinition) jen.Code {
	//
	// Generate config for enum value
	//
	//  == Example input SDL
	//
	//    """
	//    Locale describes a place with a distinct (and supported) lanugage / dialect.
	//    """
	//    enum Locale {
	//      """
	//      Canada
	//      """
	//      CA
	//      """
	//      Latin
	//      """
	//      LA @deprecated
	//    }
	//
	//  == Example output
	//
	//    &graphql.EnumValueConfig{
	//      Value:             "LA",
	//      Description:       "Latin",
	//      DeprecationReason: "No longer supported",
	//    },
	//

	desc := fetchDescription(val)
	depReason := fetchDeprecationReason(val.Directives)
	return jen.Op("&").Qual(graphqlGoPkg, "EnumValueConfig").Values(
		jen.Dict{
			jen.Id("Value"):             jen.Lit(val.Name.Value),
			jen.Id("Description"):       jen.Lit(desc),
			jen.Id("DeprecationReason"): jen.Lit(depReason),
		},
	)
}
