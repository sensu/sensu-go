package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/graphql-go/graphql/language/ast"
)

//
// Generate config for enum
//
// == Example input SDL
//
//   """
//   Locale describes a place with a distinct (and supported) lanugage / dialect.
//   """
//   enum Locale {
//     """
//     Canada
//     """
//     CA
//     """
//     Latin
//     """
//     LA @deprecated
//   }
//
// == Example output
//
//   // Locale describes a place with a distinct (and supported) lanugage / dialect.
//   type Locale string
//
//   // Locales holds enum values
//   var Locales = _EnumTypeLocaleValues{
//     CA: "CA",
//     LA: "LA",
//   }
//
//   // LocaleType describes a place with a distinct (and supported) lanugage / dialect.
//   var LocaleType = graphql.NewType("Locale", graphql.EnumKind)
//
//   // RegisterLocale registers Locale enum type with given service.
//   func RegisterLocale(svc graphql.Service) {
//     svc.RegisterEnum(_EnumTypeLocale)
//   }
//
//   // define configuration thunk
//   func _EnumTypeLocaleConfigFn() definition.EnumConfig {
//     return definition.EnumConfig{
//       Name:        "Locale",
//       Description: "Locale describes a place with a distinct (and supported) lanugage / dialect.",
//       Values: graphql.EnumValueConfigMap{
//         "CA": &graphql.EnumValueConfig{
//           Value:             "CA",
//           Description:       "Canada",
//           DeprecationReason: "",
//         },
//         "LA": &graphql.EnumValueConfig{
//           Value:             "LA",
//           Description:       "Latin",
//           DeprecationReason: "No longer supported,
//         },
//       },
//     }
//   }
//
//   // describe enums's configuration; kept private to avoid unintentional
//   // tampering at runtime.
//   var _EnumTypeLocale = graphql.EnumDesc{
//     Config: _EnumTypeConfigureLocale,
//   }
//
//   // In an attempt to avoid collisions enum values are defined within a
//   // struct that has a single public instance. The obvious tradeoff is that
//   // a developer could unintentionally mutate the values and cause heartache.
//   type _EnumTypeLocaleValues struct {
//     CA string // CA - Canada
//     LA string // LA - Latin
//   }
//
func genEnum(node *ast.EnumDefinition) jen.Code {
	code := newGroup()
	name := node.GetName().Value

	// Type description
	desc := getDescription(node)
	comment := genTypeComment(name, desc)

	// Ids
	publicValuesName := name + "s" // naive
	publicRefName := name + "Type"
	publicRefComment := genTypeComment(publicRefName, desc)
	privateEnumValuesStruct := mkPrivateID(node, "Values")
	privateConfigName := mkPrivateID(node, "Desc")
	privateConfigThunkName := mkPrivateID(node, "ConfigFn")

	//
	// Generate type that will be used to represent enum
	//
	// == Example output
	//
	//   // Locale describes a place with a distinct (and supported) lanugage / dialect.
	//   type Locale string
	//
	code.Comment(comment)
	code.Type().Id(name).String()

	//
	// Generate enum values
	//
	// == Example output
	//
	//   // Locales holds enum values
	//   var Locales = _EnumTypeLocaleValues{
	//     CA: "CA",
	//     LA: "LA",
	//   }
	//
	code.Comment(publicValuesName + " holds enum values")
	code.
		Var().Id(publicValuesName).Op("=").
		Id(privateEnumValuesStruct).
		Values(jen.DictFunc(func(d jen.Dict) {
			for _, v := range node.Values {
				d[jen.Id(v.Name.Value)] = jen.Lit(v.Name.Value)
			}
		}))

	//
	// Generate public reference to type
	//
	// == Example output
	//
	//   // Locale describes a place with a distinct (and supported) lanugage / dialect.
	//   var LocaleType = graphql.NewType("Locale", graphql.EnumKind)
	//
	code.Comment(publicRefComment)
	code.
		Var().Id(publicRefName).Op("=").
		Qual(servicePkg, "NewType").
		Call(jen.Lit(name), jen.Qual(servicePkg, "EnumKind"))

	//
	// Generate public func to register type with service
	//
	// == Example output
	//
	//   // RegisterLocale registers Locale enum type with given service.
	//   func RegisterLocale(svc graphql.Service) {
	//     svc.RegisterEnum(_EnumTypeLocaleDesc)
	//   }
	//

	code.Add(
		genRegisterFn(node, nil),
	)

	//
	// Generate type config thunk
	//
	// == Example output
	//
	//   func _EnumTypeLocaleConfigFn() definition.EnumConfig {
	//     return definition.EnumConfig{
	//       Name:        "Locale",
	//       Description: "Locale describes a place with a distinct (and supported) lanugage / dialect.",
	//       Values: graphql.EnumValueConfigMap{
	//         "CA": &graphql.EnumValueConfig{
	//           Value:             "CA",
	//           Description:       "Canada",
	//           DeprecationReason: "",
	//         },
	//         "LA": &graphql.EnumValueConfig{
	//           Value:             "LA",
	//           Description:       "Latin",
	//           DeprecationReason: "No longer supported,
	//         },
	//       },
	//     }
	//   }
	//
	code.
		Func().Id(privateConfigThunkName).
		Params().Qual(defsPkg, "EnumConfig").
		Block(
			jen.Return(jen.Qual(defsPkg, "EnumConfig").Values(jen.Dict{
				jen.Id("Description"): jen.Lit(desc),
				jen.Id("Name"):        jen.Lit(name),
				jen.Id("Values"):      genEnumValues(node.Values),
			})),
		)

	//
	// Generate type description
	//
	// == Example output
	//
	//   // describe timestamp's configuration; kept private to avoid
	//   // unintentional tampering at runtime.
	//   var _EnumTypeLocaleDesc = graphql.EnumConfig{
	//     Config: _EnumTypeLocaleConfigFn,
	//   }
	//
	code.Commentf(
		`describe %s's configuration; kept private to avoid unintentional tampering of configuration at runtime.`,
		name,
	)
	code.
		Var().Id(privateConfigName).Op("=").
		Qual(servicePkg, "EnumDesc").
		Values(jen.Dict{
			jen.Id("Config"): jen.Id(privateConfigThunkName),
		})

	//
	// Generate struct used to hold enum values
	//
	// In an attempt to avoid collisions enum values are defined within a
	// struct that has a single public instance. The obvious tradeoff is that
	// a developer could unintentionally mutate the values and cause heartache.
	//
	// == Example output
	//
	//   type _EnumTypeLocaleValues struct {
	//     CA string // CA - Canada
	//     LA string // LA - Latin
	//   }
	//
	code.
		Type().Id(privateEnumValuesStruct).
		StructFunc(func(g *jen.Group) {
			for _, v := range node.Values {
				g.Comment(v.Name.Value + " - " + getDescription(v))
				g.Id(v.Name.Value).String()
			}
		})

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
