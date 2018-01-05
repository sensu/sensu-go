package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

//
// Generates description for scalar type
//
// == Example input SDL
//
//   "Pets are the bestest family members"
//   interface Pet {
//     "name of this fine beast."
//     name: String!
//   }
//
// == Example output
//
//   // PetType - Pets are the bestest family members
//   var PetType = graphql.NewType("Pet", graphql.InterfaceKind)
//
//   // RegisterPet registers Pet interface type with given service.
//   func RegisterPet(svc graphql.Service, impl graphql.InterfaceTypeResolver) {
//     svc.RegisterInterface(_InterfaceTypePetDesc, impl)
//   }
//
//   // define configuration thunk
//   func _InterfaceTypePetConfigFn() graphql.InterfaceConfig {
//     return graphql.InterfaceConfig{
//       Name:        "Pet",
//       Description: "Pets are the bestest family members",
//       Fields:      // ...
//       ResolveType: func (_ ResolveTypeParams) string {
//         panic("Unimplemented; see PetResolver.")
//       },
//     }
//   }
//
//   // describe pet's configuration; kept private to avoid unintentional
//   // tampering at runtime.
//   var _InterfaceTypePetDesc = graphql.InterfaceDesc{
//     Config: _InterfaceTypePetConfigFn,
//   }
//
func genInterface(node *ast.InterfaceDefinition) jen.Code {
	code := newGroup()
	name := node.GetName().Value

	// Type description
	desc := getDescription(node)

	// Ids
	registerFnName := "Register" + name
	publicRefName := name + "Type"
	publicRefComment := genTypeComment(publicRefName, desc)
	privateConfigName := "_InterfaceType" + name + "Desc"
	privateConfigThunkName := "_InterfaceType" + name + "ConfigFn"

	//
	// Generate public reference to type
	//
	// == Example output
	//
	//   // PetType - Pets are the bestest family members
	//   var PetType = graphql.NewType("Pet", graphql.InterfaceKind)
	//
	code.Comment(publicRefComment)
	code.
		Var().Id(publicRefName).Op("=").
		Qual(servicePkg, "NewType").
		Call(jen.Lit(name), jen.Qual(servicePkg, "InterfaceKind"))

	//
	// Generate public func to register type with service
	//
	// == Example output
	//
	//   // RegisterPet registers Pet interface type with given service.
	//   func RegisterPet(svc graphql.Service, impl graphql.InterfaceTypeResolver) {
	//     svc.RegisterInterface(_InterfaceTypePetDesc, impl)
	//   }
	//
	code.Commentf(
		"%s registers %s interface type with given service.",
		registerFnName,
		name,
	)
	code.
		Func().Id(registerFnName).
		Params(
			jen.Id("svc").Qual(servicePkg, "Service"),
			jen.Id("impl").Qual(servicePkg, "InterfaceTypeResolver"),
		).
		Block(
			jen.Id("svc.RegisterInterface").Call(
				jen.Id(privateConfigName),
				jen.Id("impl"),
			),
		)

	//
	// Generates type config thunk
	//
	// == Example output
	//
	//   // define configuration thunk
	//   func _InterfaceTypePetConfigFn() graphql.InterfaceConfig {
	//     return graphql.InterfaceConfig{
	//       Name:        "Pet",
	//       Description: "Pets are the bestest family members",
	//       Fields:      // ...
	//       ResolveType: func (_ ResolveTypeParams) string {
	//         panic("Unimplemented; see PetResolver.")
	//       },
	//     }
	//   }
	//
	code.
		Func().Id(privateConfigThunkName).
		Params().Qual(defsPkg, "InterfaceConfig").
		Block(
			jen.Return(jen.Qual(defsPkg, "InterfaceConfig").Values(jen.Dict{
				jen.Id("Name"):        jen.Lit(name),
				jen.Id("Description"): jen.Lit(desc),
				jen.Id("Fields"):      genFields(node.Fields),
				jen.Id("ResolveType"): jen.Func().
					Params(jen.Id("_").Qual(defsPkg, "ResolveTypeParams")).
					Op("*").Qual(defsPkg, "Object").
					Block(
						jen.Comment(missingResolverNote),
						jen.Panic(jen.Lit("Unimplemented; see InterfaceTypeResolver.")),
					),
			})),
		)

	//
	// Generate type description
	//
	// == Example output
	//
	//   // ...
	//   var _InterfaceTypePetDesc = graphql.InterfaceDesc{
	//     Config: _InterfaceTypePetConfigFn,
	//   }
	//
	code.Commentf(
		`describe %s's configuration; kept private to avoid unintentional tampering of configuration at runtime.`,
		name,
	)
	code.
		Var().Id(privateConfigName).Op("=").
		Qual(servicePkg, "InterfaceDesc").
		Values(jen.Dict{
			jen.Id("Config"): jen.Id(privateConfigThunkName),
		})

	return code
}
