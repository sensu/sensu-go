package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

//
// Generate input object config
//
// == Example input SDL
//
//   """
//   SetIntervalInput sets new interval for check identified by given name.
//   """
//   input SetIntervalInput {
//     "name of the check"
//     name: String!
//
//     "new interval in seconds"
//     newInterval: Int!
//   }
//
// == Example output
//
//   // SetIntervalInput sets new interval for check identified by given name.
//   type SetIntervalInput struct {
//     // Name of the check
//     Name string
//     // NewInterval - new interval in seconds
//     NewInterval int
//   }
//
//   // SetIntervalInputType sets new interval for check identified by given name.
//   var SetIntervalInputType = graphql.NewType("Locale", graphql.InputKind)
//
//   // RegisterSetIntervalInput registers SetIntervalInput input type with given service.
//   func RegisterSetIntervalInput(svc graphql.Service) {
//     svc.RegisterInput(_InputTypeSetIntervalInput)
//   }
//
//   // expresed as thunk to ensure we always receive a new copy.
//   func _InputTypeConfigureSetIntervalInput() definition.InputObjectConfig {
//     return definition.InputObjectConfig{
//       Name: "SetIntervalInput",
//       Description: "SetIntervalInput sets a new interval for check identified by given name",
//       Fields: InputObjectFieldMap{ ... },
//    }
//
//   // describe enums's configuration; kept private to avoid unintentional
//   // tampering at runtime.
//   var _InputTypeSetInterval = graphql.InputDesc{
//     Config: _InputTypeConfigureSetIntervalInput,
//   }
//
func genInputObject(node *ast.InputObjectDefinition) jen.Code {
	code := newGroup()
	name := node.GetName().Value

	// Type description
	desc := getDescription(node)
	comment := genTypeComment(name, desc)

	// Ids
	registerFnName := "Register" + name
	publicRefName := name + "Type"
	publicRefComment := genTypeComment(publicRefName, desc)
	privateConfigName := "_InputType" + name + "Desc"
	privateConfigThunkName := "_InputType" + name + "ConfigFn"

	//
	// Generate public type
	//
	// == Example output
	//
	//   // SetIntervalInput sets new interval for check identified by given name.
	//   type SetIntervalInput struct {
	//     // Name of the check
	//     Name string
	//     // NewInterval - new interval in seconds
	//     NewInterval int
	//   }
	//
	code.Comment(comment)
	code.Type().Id(name).StructFunc(func(g *jen.Group) {
		for _, f := range node.Fields {
			g.Add(
				genInputStructField(f),
			)
		}
	})

	//
	// Generate public reference to type
	//
	// == Example output
	//
	//   // SetIntervalInputType sets new interval for check identified by given name.
	//   var SetIntervalInputType = graphql.NewType("Locale", graphql.InputKind)
	//
	code.Comment(publicRefComment)
	code.
		Var().Id(publicRefName).Op("=").
		Qual(servicePkg, "NewType").
		Call(jen.Lit(name), jen.Qual(servicePkg, "InputKind"))

	//
	// Generate public func to register type with service
	//
	// == Example output
	//
	//   // RegisterSetIntervalInput registers SetIntervalInput input type with given service.
	//   func RegisterSetIntervalInput(svc graphql.Service) {
	//     svc.RegisterInput(_InputTypeSetIntervalInputDesc)
	//   }
	//
	code.Commentf(
		"%s registers %s input type with given service.",
		registerFnName,
		name,
	)
	code.
		Func().Id(registerFnName).
		Params(jen.Id("svc").Qual(servicePkg, "Service")).
		Block(
			jen.Id("svc.RegisterInput").Call(
				jen.Id(privateConfigName),
			),
		)

	//
	// Generate type config thunk
	//
	// == Example output
	//
	//   // expresed as thunk to ensure we always receive a new copy.
	//   func _InputTypeConfigureSetIntervalInput() definition.InputObjectConfig {
	//     return definition.InputObjectConfig{
	//       Name: "SetIntervalInput",
	//       Description: "SetIntervalInput sets a new interval for check identified by given name",
	//       Fields: InputObjectFieldMap{ ... },
	//    }
	//
	code.
		Func().Id(privateConfigThunkName).
		Params().Qual(defsPkg, "InputObjectConfig").
		Block(
			jen.Return(jen.Qual(defsPkg, "InputObjectConfig").Values(jen.Dict{
				jen.Id("Name"):        jen.Lit(name),
				jen.Id("Description"): jen.Lit(desc),
				jen.Id("Fields"):      genInputObjectFields(node.Fields),
			})),
		)

	//
	// Generate type description
	//
	// == Example output
	//
	//   // describe enums's configuration; kept private to avoid unintentional
	//   // tampering at runtime.
	//   var _InputTypeSetInterval = graphql.InputDesc{
	//     Config: _InputTypeConfigureSetIntervalInput,
	//   }
	//
	code.Commentf(
		`describe %s's configuration; kept private to avoid unintentional tampering of configuration at runtime.`,
		name,
	)
	code.
		Var().Id(privateConfigName).Op("=").
		Qual(servicePkg, "InputDesc").
		Values(jen.Dict{
			jen.Id("Config"): jen.Id(privateConfigThunkName),
		})

	return code
}

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
func genInputObjectFields(fields []*ast.InputValueDefinition) jen.Code {
	return jen.Qual(defsPkg, "InputObjectConfigFieldMap").Values(
		jen.DictFunc(func(d jen.Dict) {
			for _, field := range fields {
				d[jen.Lit(field.Name.Value)] = genInputObjectField(field)
			}
		}),
	)
}

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
func genInputObjectField(field *ast.InputValueDefinition) jen.Code {
	return jen.Op("&").Qual(defsPkg, "InputObjectFieldConfig").Values(jen.Dict{
		jen.Id("Type"):         genInputTypeReference(field.Type),
		jen.Id("Description"):  genDescription(field),
		jen.Id("DefaultValue"): genValue(field.DefaultValue),
	})
}

//
// Generate input struct field
//
// == Example input SDL
//
//   """
//   SetIntervalInput sets new interval for check identified by given name.
//   """
//   input ConfigureIntervalInput {
//     "name of the check"
//     name: String!
//
//     "new interval in seconds"
//     newInterval: Int!
//   }
//
// == Example output
//
//   // SetIntervalInput sets new interval for check identified by given name.
//   type SetIntervalInput struct {
//     // Name of the check
//     Name string
//     // NewInterval - new interval in seconds
//     NewInterval int
//   }
//
func genInputStructField(f *ast.InputValueDefinition) jen.Code {
	name := toFieldName(f.Name.Value)
	desc := getDescription(f)
	depr := getDeprecationReason(f.Directives)
	tRef := genConcreteTypeReference(f.Type)
	comment := genFieldComment(name, desc, depr)

	return jen.Comment(comment).Id(name).Add(tRef)
}
