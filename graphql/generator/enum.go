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

	// Type description
	desc := getDescription(node)
	comment := genTypeComment(name, desc)

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
	code.Comment(comment)
	code.Func().Id(name).Params().Qual(defsPkg, "EnumConfig").Block(
		jen.Return(jen.Qual(defsPkg, "EnumConfig").Values(jen.Dict{
			jen.Id("Description"): jen.Lit(desc),
			jen.Id("Name"):        jen.Lit(name),
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

	return jen.Qual(defsPkg, "EnumValueConfigMap").Values(
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

	return jen.Op("&").Qual(defsPkg, "EnumValueConfig").Values(
		jen.Dict{
			jen.Id("DeprecationReason"): genDeprecationReason(val.Directives),
			jen.Id("Description"):       genDescription(val),
			jen.Id("Value"):             jen.Lit(val.Name.Value),
		},
	)
}
